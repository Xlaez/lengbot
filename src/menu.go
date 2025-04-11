package src

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func SendMenu(bot *tgbotapi.BotAPI, chatId int64) {
	msg := tgbotapi.NewMessage(chatId, "🎮 Welcome to One-Minute Trivia Wars!\nChoose an option:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎲 Start Random Match", "start_match"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Select Category", "select_category"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧑‍🤝‍🧑 1v1 Challenge", "challenge_mode"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏅 Leaderboard", "show_rank"),
			tgbotapi.NewInlineKeyboardButtonData("🌍 Global Stats", "show_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🕒 Timed Game", "select_time"),
		),
	)
	bot.Send(msg)
}

func SendTimeModeMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "⏳ How long should the trivia game last?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("2 Minutes", "timed_2"),
			tgbotapi.NewInlineKeyboardButtonData("5 Minutes", "timed_5"),
			tgbotapi.NewInlineKeyboardButtonData("10 Minutes", "timed_10"),
		),
	)
	bot.Send(msg)
}
