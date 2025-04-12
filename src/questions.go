package src

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Question struct {
	Text     string `json:"text"`
	Answer   string `json:"answer"`
	Category string `json:"category"`
}

var questions []Question

func LoadQuestions() {
	filepath := filepath.Join("temp", "questions.json")
	data, err := os.ReadFile(filepath)

	if err != nil {
		log.Fatal("Failed to load questions:", err)
	}
	json.Unmarshal(data, &questions)
}

func SendNextQuestion(bot *tgbotapi.BotAPI, gameID string) {
	game := ActiveGames[gameID]

	if game.CurrentQ >= len(questions) {
		EndGame(bot, gameID)
		return
	}

	q := questions[game.CurrentQ]
	game.CurrentQ++

	bot.Send(tgbotapi.NewMessage(game.Player1, fmt.Sprintf("❓ Question: %s", q.Text)))
	bot.Send(tgbotapi.NewMessage(game.Player2, fmt.Sprintf("❓ Question: %s", q.Text)))
}