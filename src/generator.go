package src

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Xlaez/lengbot/configs"
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
		category = "music"
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
	Questions = append(Questions, text)

	if IsDuplicateQuestion(text) {
		return GenerateTriviaQuestion(category)
	}


	return text, nil
}

func getPromptForCategory(cat string) string {
	basePrompt := `
Generate a trivia question with 4 options and the correct answer.
Do not repeat questions again.
Never repeat any questions, they should be unique.
Use this format exactly:

Question: ...
Options:
A. ...
B. ...
C. ...
D. ...
Answer: A
`

	switch strings.ToLower(cat) {
	case "science":
		return basePrompt + "\nCategory: Science"
	case "music":
		return basePrompt + "\nCategory: Music"
	case "football":
		return basePrompt + "\nCategory: Football"
	case "arts":
		return basePrompt + "\nCategory: Arts"
	case "tech":
		return basePrompt + "\nCategory: Technology"
	case "africa":
		return basePrompt + "\nCategory: African History or Culture"
	default:
		return basePrompt + "\nCategory: General Knowledge"
	}
}

// func GenerateTriviaQuestion(category string,) (question, answer string, err error){

// 	ctx := context.Background()

// 	req := openai.ChatCompletionRequest{
// 		Model: openai.GPT3Dot5Turbo,
// 		Messages: []openai.ChatCompletionMessage{
// 			{
// 				Role:    openai.ChatMessageRoleSystem,
// 				Content: fmt.Sprintf("You are a trivia master. Generate trivia questions and answers on the subject of %.", category),
// 			},
// 			{
// 				Role:    openai.ChatMessageRoleUser,
// 				Content: `Generate one trivia question and its correct answer in the format:
// Question: ...
// Answer: ...`,
// 			},
// 		},
// 		Temperature: 0.7,
// 	}

// 	resp, err := client.CreateChatCompletion(ctx, req)
// 	if err != nil {
// 		return "", "", err
// 	}

// 	content := resp.Choices[0].Message.Content
// 	lines := strings.Split(content, "\n")

// 	if len(lines) < 2 {
// 		return "", "", errors.New("unexpected response from OpenAI")
// 	}

// 	question = strings.TrimPrefix(lines[0], "Question: ")
// 	answer = strings.TrimPrefix(lines[1], "Answer: ")

// 	return strings.TrimSpace(question), strings.TrimSpace(answer), nil
// }
