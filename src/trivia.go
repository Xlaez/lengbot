package src

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TriviaSession struct {
	Player1 int64 
	Player2 int64
	Scores map[int64]int
	Questions []Question
	CurrentQ int
	IsActive bool
	EndsAt time.Time
}

var waitingPool = make(map[string]*tgbotapi.User)
var activeGames = make(map[string]*TriviaSession)

func StartTriviaMatch(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	// Handle matching of opponent
	// Handle edge-cases

	key := "random"

	if waitingPool[key] == nil {
		waitingPool[key] = msg.From
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Waiting for an opponent..."))
		return
	}

	if waitingPool[key].ID == msg.From.ID {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "You're already waiting for a match!"))
		return
	}

	// Match found
	player1 := waitingPool[key]
	player2 := msg.From
	waitingPool[key] = nil

	// GameID for session is a concatenation of players IDs 
	gameID := fmt.Sprintf("%d_%d", player1.ID, player2.ID)

	activeGames[gameID] = &TriviaSession{
		Player1:  player1.ID,
		Player2:  player2.ID,
		Scores:   map[int64]int{player1.ID: 0, player2.ID: 0},
		CurrentQ: 0,
		IsActive: true,
	}

	bot.Send(tgbotapi.NewMessage(player1.ID, fmt.Sprintf("🎮 You're matched with %s! Get ready!", player2.FirstName)))
	bot.Send(tgbotapi.NewMessage(player2.ID, fmt.Sprintf("🎮 You're matched with %s! Get ready!", player1.FirstName)))

	SendNextQuestion(bot, gameID)
}

func ProcessTriviaAnswer(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, category string) {

	if category == "" {
		category = "science"
	}

	for id, game := range activeGames {
		if !game.IsActive {
			continue
		}
		if msg.From.ID != game.Player1 && msg.From.ID != game.Player2 {
			continue
		}

		gameId := fmt.Sprintf("%d_%d",game.Player1, game.Player2)
		

		// Current question
		// q := questions[game.CurrentQ-1]
		// if msg.Text == "" {
		// 	return
		// }

		// // Has already answered?
		// if game.Scores[msg.From.ID] >= 3 {
		// 	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "You already won 😄"))
		// 	return
		// }


		answer := CurrentAnswer[gameId]
		if Normalize(msg.Text) == Normalize(answer) {
			game.Scores[msg.From.ID]++
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "✅ Correct!"))


			// Check if someone won

			// Old logic for when it was length
			// if game.Scores[msg.From.ID] >= 5 {
			// 	EndGame(bot, id)
			// 	return
			// }
			
			delete(CurrentAnswer, gameId)

			sendNextAIQuestion(bot, gameId, category)

			if time.Now().After(game.EndsAt) {
				EndGame(bot, id)
				return
			}

			SendNextQuestion(bot ,id)
		} else {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Nope! Try again"))
		}
	}
}

func EndGame(bot *tgbotapi.BotAPI, gameID string) {
	game := activeGames[gameID]
	game.IsActive = false

	var winnerID int64
	if game.Scores[game.Player1] > game.Scores[game.Player2] {
		winnerID = game.Player1
	} else {
		winnerID = game.Player2
	}

	IncrementWins(winnerID)

	bot.Send(tgbotapi.NewMessage(game.Player1, fmt.Sprintf("🏆 Game Over! %d wins!", winnerID)))
	bot.Send(tgbotapi.NewMessage(game.Player2, fmt.Sprintf("🏆 Game Over! %d wins!", winnerID)))

	delete(CurrentAnswer, gameID)
	delete(activeGames, gameID)
}

func StartTriviaMatchWithCategory(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, category string) {
	key := category

	if waitingPool[key] == nil {
		waitingPool[key] = msg.From
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Waiting for an opponent in %s category...", category)))
		return
	}

	if waitingPool[key].ID == msg.From.ID {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "You're already waiting for a match!"))
		return
	}

	player1 := waitingPool[key]
	player2 := msg.From
	waitingPool[key] = nil

	filteredQuestions := FilterQuestionsByCategory(category)
	if len(filteredQuestions) == 0 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "No questions available for this category!"))
		return
	}

	gameID := fmt.Sprintf("%d_%d", player1.ID, player2.ID)
	activeGames[gameID] = &TriviaSession{
		Player1:   player1.ID,
		Player2:   player2.ID,
		Scores:    map[int64]int{player1.ID: 0, player2.ID: 0},
		CurrentQ:  0,
		IsActive:  true,
		Questions: filteredQuestions,
	}

	bot.Send(tgbotapi.NewMessage(player1.ID, fmt.Sprintf("🎮 You're matched with %s in %s!", player2.FirstName, category)))
	bot.Send(tgbotapi.NewMessage(player2.ID, fmt.Sprintf("🎮 You're matched with %s in %s!", player1.FirstName, category)))

	SendNextQuestion(bot, gameID)
}

func Start1v1Challenge(bot *tgbotapi.BotAPI, challengerID, opponentID int64) {
	if challengerID == opponentID {
		bot.Send(tgbotapi.NewMessage(opponentID, "You can't challenge yourself 😅"))
		return
	}

	gameID := fmt.Sprintf("%d_%d", challengerID, opponentID)

	filteredQuestions := questions

	if len(filteredQuestions) == 0 {
		bot.Send(tgbotapi.NewMessage(opponentID, "No trivia questions available 😢"))
		return
	}

	activeGames[gameID] = &TriviaSession{
		Player1:   challengerID,
		Player2:   opponentID,
		Scores:    map[int64]int{challengerID: 0, opponentID: 0},
		CurrentQ:  0,
		IsActive:  true,
		Questions: filteredQuestions,
	}

	bot.Send(tgbotapi.NewMessage(challengerID, fmt.Sprintf("🎮 Your challenge has been accepted by %d! Game on!", opponentID)))
	bot.Send(tgbotapi.NewMessage(opponentID, fmt.Sprintf("🎮 You've accepted a challenge from %d! Get ready!", challengerID)))

	SendNextQuestion(bot, gameID)
}
