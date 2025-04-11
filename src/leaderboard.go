package src

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var leaderboardFile = "temp/leaderboard.json"
var leaderboard = map[int64]int{}
var leaderboardLock sync.Mutex


func LoadLeaderboard() {
	file, err := os.ReadFile(leaderboardFile)
	if err != nil {
		fmt.Println("⚠️ No leaderboard record found, starting fresh.")
		return
	}
	json.Unmarshal(file, &leaderboard)
}


func SaveLeaderboard() {
	leaderboardLock.Lock()
	defer leaderboardLock.Unlock()

	data, _ := json.MarshalIndent(leaderboard, "", "  ")
	os.WriteFile(leaderboardFile, data, 0644)
}

func IncrementWins(userID int64) {
	leaderboardLock.Lock()
	defer leaderboardLock.Unlock()

	leaderboard[userID]++
	SaveLeaderboard()
}

func SendLeaderboard(bot *tgbotapi.BotAPI, chatID int64) {
	type entry struct {
		ID    int64
		Score int
	}

	var entries []entry
	for k, v := range leaderboard {
		entries = append(entries, entry{k, v})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	msg := "🏅 <b>Trivia Leaderboard</b>\n\n"
	max := 5
	if len(entries) < max {
		max = len(entries)
	}
	for i := 0; i < max; i++ {
		userID := entries[i].ID
		score := entries[i].Score
		msg += fmt.Sprintf("%d. 🧠 <code>%d</code> - %d wins\n", i+1, userID, score)
	}

	if max == 0 {
		msg += "No players have won yet. Be the first! 🚀"
	}

	message := tgbotapi.NewMessage(chatID, msg)
	message.ParseMode = "HTML"
	bot.Send(message)
}

