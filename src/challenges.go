package src

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Challenge struct {
	ChallengerID int64
	OpponentID   int64
	Category     string 
}

var pendingChallenges = make(map[string]Challenge)
// key = gameID
var CurrentAnswer = make(map[string]string)

func AskForChallenge(bot *tgbotapi.BotAPI, chatID int64, category string) {
	caser := cases.Title(language.English)
	title := caser.String(category)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔗 Send this link to your friend to challenge them in %s:", title))

	startData := fmt.Sprintf("challenge_%d_%s", chatID, category)
	link := fmt.Sprintf("https://t.me/%s?start=%s", bot.Self.UserName, startData)

	btn := tgbotapi.NewInlineKeyboardButtonURL("Accept Challenge", link)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)

	bot.Send(msg)
}


func SendChallengeCategoryMenu(bot *tgbotapi.BotAPI, chatId int64) {
	msg := tgbotapi.NewMessage(chatId, "Pick a category for the 1v1 challenge:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎶 Music", "challenge_music"),
			tgbotapi.NewInlineKeyboardButtonData("📘 Arts", "challenge_arts"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚽ Football", "challenge_football"),
			tgbotapi.NewInlineKeyboardButtonData("🔬 Science", "challenge_science"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌍 Africa", "category_africa"),
			tgbotapi.NewInlineKeyboardButtonData("💻 Tech", "category_tech"),
		),
	)
	bot.Send(msg)
}

func Start1v1ChallengeWithCategory(bot *tgbotapi.BotAPI, challengerID, opponentID int64, category string) {
	if challengerID == opponentID {
		bot.Send(tgbotapi.NewMessage(opponentID, "You can't challenge yourself 😅"))
		return
	}

	gameID := fmt.Sprintf("%d_%d", challengerID, opponentID)
	ActiveGames[gameID] = &TriviaSession{
		Player1:   challengerID,
		Player2:   opponentID,
		Scores:    map[int64]int{challengerID: 0, opponentID: 0},
		CurrentQ:  0,
		IsActive:  true,
		// Questions: filteredQuestions,
	}

	bot.Send(tgbotapi.NewMessage(challengerID, fmt.Sprintf("🎯 %d accepted your %s challenge!", opponentID, category)))
	bot.Send(tgbotapi.NewMessage(opponentID, fmt.Sprintf("✅ You joined a %s challenge from %d!", category, challengerID)))

	msg := tgbotapi.NewMessage(challengerID, "⏳ How long should the match last?")
	msg.ReplyMarkup = durationButtons(gameID)
	bot.Send(msg)
}

func durationButtons(gameID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("2 min", "duration_"+gameID+"_2"),
			tgbotapi.NewInlineKeyboardButtonData("5 min", "duration_"+gameID+"_5"),
			tgbotapi.NewInlineKeyboardButtonData("10 min", "duration_"+gameID+"_10"),
		),
	)
}


func StartTimedGame(bot *tgbotapi.BotAPI, gameID, category string, minutes int) {
	game := ActiveGames[gameID]
	if game == nil {
		return
	}

	game.EndsAt = time.Now().Add(time.Duration(minutes) * time.Minute)
	game.IsActive = true

	// Send message and start first question
	bot.Send(tgbotapi.NewMessage(game.Player1, fmt.Sprintf("Game starts now! Duration: %d minutes", minutes)))
	bot.Send(tgbotapi.NewMessage(game.Player2, fmt.Sprintf("Game starts now! Duration: %d minutes", minutes)))

	SendNextAIQuestion(bot, gameID, category)

	// Schedule end
	go func() {
		time.Sleep(time.Duration(minutes) * time.Minute)
		EndGame(bot, gameID)
	}()
}

func SendNextAIQuestion(bot *tgbotapi.BotAPI, gameID, category string) {
	game := ActiveGames[gameID]
	if game == nil {
		return
	}

	// Get usernames for both players
	// player1, _ := bot.GetUserProfilePhotos(tgbotapi.NewUserProfilePhotos(game.Player1))
	// player2, _ := bot.GetUserProfilePhotos(tgbotapi.NewUserProfilePhotos(game.Player2))

	// Default fallback
	player1Username := "Player 1"
	player2Username := "Player 2"


	// if player1 != nil {
		// player1Username = player1
	// }

	// if player2 != nil {
		// player2Username = player2.UserName
	// }

	scoreMessage := fmt.Sprintf("🎮 Scores:\n%s: %d\n%s: %d\n\n",
		player1Username, game.Scores[game.Player1],
		player2Username, game.Scores[game.Player2])

	// Send scores to both players
	bot.Send(tgbotapi.NewMessage(game.Player1, scoreMessage))
	bot.Send(tgbotapi.NewMessage(game.Player2, scoreMessage))


	WrongAnswersThisRound[gameID] =  make(map[int64]bool)
	CorrectAnswersThisRound[gameID] = make(map[int64]bool)

	text,err := GenerateTriviaQuestion(category)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(game.Player1, "⚠️ Failed to generate a question."))
		bot.Send(tgbotapi.NewMessage(game.Player2, "⚠️ Failed to generate a question."))
		EndGame(bot, gameID)
		return
	}

	qa, err := ParseMCQText(text)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(game.Player1, "❌ Invalid question format"))
		bot.Send(tgbotapi.NewMessage(game.Player2, "❌ Invalid question format"))
		EndGame(bot, gameID)
		return
	}

	// Store answer to verify later
	CurrentAnswer[gameID] = qa.Answer

	AnsweredThisRound[gameID] = make(map[int64]bool)

	// Format question with options
	qText := fmt.Sprintf("❓ %s\n\nA. %s\nB. %s\nC. %s\nD. %s",
		qa.Question,
		qa.Options["A"],
		qa.Options["B"],
		qa.Options["C"],
		qa.Options["D"],
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("A", "answer_A_"+gameID),
			tgbotapi.NewInlineKeyboardButtonData("B", "answer_B_"+gameID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("C", "answer_C_"+gameID),
			tgbotapi.NewInlineKeyboardButtonData("D", "answer_D_"+gameID),
		),
	)

	// Send to both players
	msg1 := tgbotapi.NewMessage(game.Player1, qText)
	msg1.ReplyMarkup = keyboard
	bot.Send(msg1)

	msg2 := tgbotapi.NewMessage(game.Player2, qText)
	msg2.ReplyMarkup = keyboard
	bot.Send(msg2)
}
