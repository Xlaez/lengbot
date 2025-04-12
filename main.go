package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Xlaez/lengbot/configs"
	"github.com/Xlaez/lengbot/src"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot * tgbotapi.BotAPI
	Category = "general"
)

func main () {
	cfg := configs.GetConfig()

	var err error

	bot, err = tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Panic(err.Error())
	}

	src.LoadLeaderboard()

	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	updates := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, updates)

	log.Println("Start listening for updates. Press enter to stop")

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
			case <- ctx.Done():
				return
			case update := <- updates: 
				handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(update.Message)
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)
	}
}

func handleMessage(message *tgbotapi.Message){
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	log.Printf("%s wrote %s", user.FirstName, text)

	var err error

	if strings.HasPrefix(text, "/start challenge_") {
		parts := strings.Split(strings.TrimPrefix(text, "/start challenge_"), "_")
		if len(parts) != 2 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid challenge link."))
			return
		}
		challengerID, err := strconv.ParseInt(parts[0], 10, 64)
		Category = parts[1]

		if err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid challenge ID."))
			return
		}

		src.Start1v1ChallengeWithCategory(bot, challengerID, message.From.ID, Category)
		return
	}

	switch text {
	case "/trivia":
		src.StartTriviaMatch(bot, message)
	case "/start":
		src.SendMenu(bot, message.Chat.ID)
	case "/rank":
		src.SendLeaderboard(bot, message.Chat.ID)
	default:
		src.ProcessTriviaAnswer(bot, message, Category)
	}

	if err != nil {
		log.Printf("An error occurred: %s", err.Error())
	}
}

func handleButton(query *tgbotapi.CallbackQuery) {
	data := query.Data
	chatID := query.Message.Chat.ID

	if strings.HasPrefix(data, "duration_") {
		parts := strings.Split(data, "_")
		gameID := parts[1] + "_" + parts[2] // player1_player2
		minutes, _ := strconv.Atoi(parts[3])
		src.StartTimedGame(bot, gameID, Category , minutes)
	}

	if strings.HasPrefix(query.Data, "answer_") {
		// Process the answer
		parts := strings.Split(data, "_")
		playerAnswer := parts[1]
		gameID := parts[2] + "_" + parts[3]
		game, ok := src.ActiveGames[gameID]
		if !ok || !game.IsActive {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Game not active or found"))
			return
		}

		userID := query.From.ID

		// Track if a player has answered
		if src.AnsweredThisRound[gameID][userID] {
			bot.Send(tgbotapi.NewMessage(chatID, "🕓 You already answered this one!"))
			return
		}
		src.AnsweredThisRound[gameID][userID] = true


		// Check the answer
		correct := src.CurrentAnswer[gameID]
		if src.Normalize(playerAnswer) == src.Normalize(correct) {
				game.Scores[userID]++
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Correct!"))
				src.CorrectAnswersThisRound[gameID][userID] = true
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Wrong! Try again next one."))

			src.WrongAnswersThisRound[gameID][userID] = true
		}

		// If both players have answered, proceed to the next question
		if (len(src.CorrectAnswersThisRound[gameID]) > 0) || len(src.WrongAnswersThisRound[gameID]) == 2{
			// If both players have answered, proceed with another
			src.SendNextAIQuestion(bot, gameID, Category)
		}
	}


	switch data {
	case "start_match":
		src.StartTriviaMatch(bot, query.Message)
	case "select_category":
		src.SendCategoryMenu(bot ,chatID)
	case "show_rank":
		src.SendLeaderboard(bot, chatID)
	// case "show_stats":
	// 	src.SendGlobalStats(chatID)
	case "challenge_mode":
		src.SendChallengeCategoryMenu(bot, chatID)
	
	// Categories

	case "category_music":
		src.StartTriviaMatchWithCategory(bot, query.Message, "music")
	case "category_arts":
		src.StartTriviaMatchWithCategory(bot, query.Message, "arts")
	case "category_africa":
		src.StartTriviaMatchWithCategory(bot, query.Message, "africa")
	case "category_science":
		src.StartTriviaMatchWithCategory(bot, query.Message, "science")
	case "category_football":
		src.StartTriviaMatchWithCategory(bot, query.Message, "football")
	case "category_tech":
		src.StartTriviaMatchWithCategory(bot, query.Message, "tech")

	// Challenges

	case "challenge_music":
		src.AskForChallenge(bot, query.From.ID, "music")
	case "challenge_arts":
		src.AskForChallenge(bot, query.From.ID, "arts")
	case "challenge_africa":
		src.AskForChallenge(bot, query.From.ID, "africa")
	case "challenge_science":
		src.AskForChallenge(bot, query.From.ID, "science")
	case "challenge_football":
		src.AskForChallenge(bot, query.From.ID, "football")
	case "challenge_tech":
		src.AskForChallenge(bot, query.Message.From.ID, "tech")
	}

	// Acknowledge callback
	bot.Send(tgbotapi.NewCallback(query.ID, ""))
}


