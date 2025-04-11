package src

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendCategoryMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose a trivia category:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎶 Music", "category_music"),
			tgbotapi.NewInlineKeyboardButtonData("📘 Arts", "category_arts"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚽ Football", "category_football"),
			tgbotapi.NewInlineKeyboardButtonData("🔬 Science", "category_science"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌍 Africa", "category_africa"),
			tgbotapi.NewInlineKeyboardButtonData("💻 Tech", "category_tech"),
		),
	)
	bot.Send(msg)
}
