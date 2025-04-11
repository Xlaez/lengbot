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

func GenerateTriviaQuestion(category string) (string, string, error) {

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



	fmt.Println("⚡ RAW HF BODY:", string(resp.Body()))

	if err != nil {
		return "", "", err
	}

	var result []struct {
		GeneratedText string `json:"generated_text"`
	}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil || len(result) == 0 {
		return "", "", fmt.Errorf("failed to parse Hugging Face response: %v", err)
	}

	text := result[0].GeneratedText
	fmt.Println("✅ Parsed HF response:", text)

	// Extract Q & A from response text
	if strings.Contains(text, "Answer:") && strings.Contains(text, "Question:") {
		lines := strings.Split(text, "\n")
		var q, a string
		for _, line := range lines {
			if strings.Contains(line, "Question:") {
				q = strings.TrimSpace(strings.TrimPrefix(line, "Question:"))
			}
			if strings.Contains(line, "Answer:") {
				a = strings.TrimSpace(strings.TrimPrefix(line, "Answer:"))
			}
		}
		if q != "" && a != "" {
			return q, a, nil
		}
	}

	return strings.TrimSpace(text), "I don't know", nil
}

func getPromptForCategory(cat string) string {
	examples := `
Write a trivia question and its correct answer.

Example 1:
Question: What is the tallest mountain in the world?
Answer: Mount Everest

Example 2:
Question: Who painted the Mona Lisa?
Answer: Leonardo da Vinci

Now your turn:
`

	switch strings.ToLower(cat) {
	case "science":
		return examples + "\nGenerate a science trivia question and answer."
	case "football":
		return examples + "\nGenerate a football trivia question and answer."
	case "music":
		return examples + "\nGenerate a music-related trivia question and answer."
	case "arts":
		return examples + "\nGenerate an arts-related trivia question and answer."
	case "tech":
		return examples + "\nGenerate a technology-related trivia question and answer."
	case "africa":
		return examples + "\nGenerate a trivia question and answer related to African history or culture."
	default:
		return examples + "\nGenerate a general knowledge trivia question and answer."
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
