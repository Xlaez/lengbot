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
var CurrentAnswer = make(map[string]string) // key = gameID

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
		// Add more as needed
	)
	bot.Send(msg)
}

func Start1v1ChallengeWithCategory(bot *tgbotapi.BotAPI, challengerID, opponentID int64, category string) {
	if challengerID == opponentID {
		bot.Send(tgbotapi.NewMessage(opponentID, "You can't challenge yourself 😅"))
		return
	}

	filteredQuestions := FilterQuestionsByCategory(category)
	if len(filteredQuestions) == 0 {
		bot.Send(tgbotapi.NewMessage(opponentID, fmt.Sprintf("No questions available for %s ❌", category)))
		return
	}

	gameID := fmt.Sprintf("%d_%d", challengerID, opponentID)
	activeGames[gameID] = &TriviaSession{
		Player1:   challengerID,
		Player2:   opponentID,
		Scores:    map[int64]int{challengerID: 0, opponentID: 0},
		CurrentQ:  0,
		IsActive:  true,
		Questions: filteredQuestions,
	}

	bot.Send(tgbotapi.NewMessage(challengerID, fmt.Sprintf("🎯 %d accepted your %s challenge!", opponentID, category)))
	bot.Send(tgbotapi.NewMessage(opponentID, fmt.Sprintf("✅ You joined a %s challenge from %d!", category, challengerID)))

	msg := tgbotapi.NewMessage(challengerID, "⏳ How long should the match last?")
	msg.ReplyMarkup = durationButtons(gameID)
	bot.Send(msg)


	SendNextQuestion(bot, gameID)
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
	game := activeGames[gameID]
	if game == nil {
		return
	}

	game.EndsAt = time.Now().Add(time.Duration(minutes) * time.Minute)
	game.IsActive = true

	// Send message and start first question
	bot.Send(tgbotapi.NewMessage(game.Player1, fmt.Sprintf("Game starts now! Duration: %d minutes", minutes)))
	bot.Send(tgbotapi.NewMessage(game.Player2, fmt.Sprintf("Game starts now! Duration: %d minutes", minutes)))

	sendNextAIQuestion(bot, gameID, category)

	// Schedule end
	go func() {
		time.Sleep(time.Duration(minutes) * time.Minute)
		EndGame(bot, gameID)
	}()
}

func sendNextAIQuestion(bot *tgbotapi.BotAPI, gameID, category string) {
	game := activeGames[gameID]
	q, a, err := GenerateTriviaQuestion(category)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(game.Player1, "⚠️ Failed to generate a question."))
		bot.Send(tgbotapi.NewMessage(game.Player2, "⚠️ Failed to generate a question."))
		EndGame(bot, gameID)
		return
	}

	// Store answer somewhere for verification
	CurrentAnswer[gameID] = a

	bot.Send(tgbotapi.NewMessage(game.Player1, "❓ "+q))
	bot.Send(tgbotapi.NewMessage(game.Player2, "❓ "+q))
}

