package src

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/Xlaez/lengbot/configs"
	"github.com/Xlaez/lengbot/src/enums"
	"github.com/go-resty/resty/v2"
)


var hfClient = resty.New()


type HFResponse struct {
	Answer string  `json:"answer"`
	Score  float64 `json:"score"`
	Start  int     `json:"start"`
	End    int     `json:"end"`
}

type TextGenPayload struct {
	Inputs string `json:"inputs"`
}

type HFTextGenResponse []struct {
	GeneratedText string `json:"generated_text"`
}

var Questions []string = make([]string, 0)

func GenerateTriviaQuestion(category string) (string, error) {

	if category == "" {
		category = enums.General
	}

	prompt := getPromptForCategory(category)

	config := configs.GetConfig()

	resp, err := hfClient.R().
		SetHeader("Authorization", "Bearer "+ config.HuggingFaceApiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"inputs": prompt}).
		// Post("https://api-inference.huggingface.co/models/HuggingFaceH4/zephyr-7b-beta")
		// Post("https://api-inference.huggingface.co/models/google/flan-t5-base")
		Post("https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.1")

	if err != nil {
		return "", err
	}

	fmt.Println("⚡ RAW HF BODY:", string(resp.Body()))
	
	var result []struct {
		GeneratedText string `json:"generated_text"`
	}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil || len(result) == 0 {
		return "",  fmt.Errorf("failed to parse Hugging Face response: %v", err)
	}

	text := strings.TrimSpace(result[0].GeneratedText)

	if IsDuplicateQuestion(text) {
		return GenerateTriviaQuestion(category  + " " + fmt.Sprintf("unique_%d", rand.Intn(1000)))
	}


	return text, nil
}

func getPromptForCategory(cat string) string {
	basePrompt := `
Generate a completely new and unique trivia question with 4 options and the correct answer.
The question MUST be different from any previous questions.
Make it challenging but fair.
Use this format exactly:

Question: ...
Options:
A. ...
B. ...
C. ...
D. ...
Answer: A
`

	difficultyLevels := []string{"easy", "medium", "challenging", "difficult"}
	randomDifficulty := difficultyLevels[rand.Intn(len(difficultyLevels))]

	switch strings.ToLower(cat) {
	case enums.Science:
		return basePrompt + fmt.Sprintf("\nCategory: Science\nDifficulty: %s", randomDifficulty)
	case enums.Music:
		return basePrompt + fmt.Sprintf("\nCategory: Music\nDifficulty: %s", randomDifficulty)
	case enums.Football:
		return basePrompt + fmt.Sprintf("\nCategory: Soccer\nDifficulty: %s", randomDifficulty)
	case enums.Arts:
		return basePrompt + fmt.Sprintf("\nCategory: Arts\nDifficulty: %s", randomDifficulty)
	case enums.Tech:
		return basePrompt + fmt.Sprintf("\nCategory: Technology\nDifficulty: %s", randomDifficulty)
	case enums.Africa:
		return basePrompt + fmt.Sprintf("\nCategory: African History or Culture\nDifficulty: %s", randomDifficulty)
	default:
		return basePrompt + fmt.Sprintf("\nCategory: General Knowledge\nDifficulty: %s", randomDifficulty)
	}
}