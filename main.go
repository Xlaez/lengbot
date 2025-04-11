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
	firstMenu = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	nextButton = "Next"
	backButton = "Back"
	tutorialButton = "Tutorial"

	screaming = false
	bot * tgbotapi.BotAPI

	firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(backButton, nextButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://dolphjs.com"),
		),
	)
)

func main () {
	cfg := configs.GetConfig()

	var err error

	bot, err = tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Panic(err.Error())
	}

	src.LoadQuestions()
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

	// if strings.HasPrefix(text, "/start challenge_") {
	// 	challengerIDStr := strings.TrimPrefix(text, "/start challenge_")
	// 	challengerID, err := strconv.ParseInt(challengerIDStr, 10, 64)
	// if err != nil {
	// 	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid challenge link 😕"))
	// 	return
	// }

	// 	src.Start1v1Challenge(bot, challengerID, message.From.ID)
	// 	return
	// }

	var category string
	
	if strings.HasPrefix(text, "/start challenge_") {
		parts := strings.Split(strings.TrimPrefix(text, "/start challenge_"), "_")
	if len(parts) != 2 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid challenge link."))
		return
	}
	challengerID, err := strconv.ParseInt(parts[0], 10, 64)
	category = parts[1]

	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid challenge ID."))
		return
	}

		src.Start1v1ChallengeWithCategory(bot, challengerID, message.From.ID, category)
		return
	}

	switch text {
	case "/trivia":
		src.StartTriviaMatch(bot, message)
	case "/start":
		src.SendMenu(bot, message.Chat.ID)
	case "/rank":
		src.SendLeaderboard(bot, message.Chat.ID)
	case "/aiq":
	q, a, err := src.GenerateTriviaQuestion(category)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Error: " + err.Error()))
	} else {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❓ " + q + "\n✅ " + a))
	}

	default:
		src.ProcessTriviaAnswer(bot, message, category)
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
		src.StartTimedGame(bot, gameID, "football" , minutes)
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


