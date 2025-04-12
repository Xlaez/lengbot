package src

import (
	"fmt"
	"strings"
)

func Normalize(s string) string {
	// Normalize by stripping the prefix (e.g., "A.", "B.", "C.", "D.") and trimming spaces
	s = strings.ToLower(strings.TrimSpace(s))
	
	// If the string starts with "A.", "B.", "C." or "D.", remove the prefix (e.g., "A.", "B.")
	if len(s) > 2 && (s[1] == '.') {
		s = string(s[0])
	}
	return s
}



func FilterQuestionsByCategory(category string) []Question {
	var filtered []Question
	for _, q := range questions {
		if strings.ToLower(q.Category) == strings.ToLower(category) {
			filtered = append(filtered, q)
		}
	}
	return filtered
}

func ParseMCQText(text string) (TriviaQA, error) {
	lines := strings.Split(text, "\n")
	qa := TriviaQA{Options: make(map[string]string)}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Question:"):
			qa.Question = strings.TrimSpace(strings.TrimPrefix(line, "Question:"))
		case strings.HasPrefix(line, "A."):
			qa.Options["A"] = strings.TrimSpace(strings.TrimPrefix(line, "A."))
		case strings.HasPrefix(line, "B."):
			qa.Options["B"] = strings.TrimSpace(strings.TrimPrefix(line, "B."))
		case strings.HasPrefix(line, "C."):
			qa.Options["C"] = strings.TrimSpace(strings.TrimPrefix(line, "C."))
		case strings.HasPrefix(line, "D."):
			qa.Options["D"] = strings.TrimSpace(strings.TrimPrefix(line, "D."))
		case strings.HasPrefix(line, "Answer:"):
			qa.Answer = strings.TrimSpace(strings.TrimPrefix(line, "Answer:"))
		}
	}

	if qa.Question == "" || len(qa.Options) != 4 || qa.Answer == "" {
		return qa, fmt.Errorf("incomplete MCQ format: %+v", qa)
	}
	return qa, nil
}

func IsDuplicateQuestion(question string) bool {
	// Logic to check if the question was already asked

	for _, game := range ActiveGames {
		// Loop through active games and check if the question is already asked
		for _, q := range game.Questions {
			if q == question {
				return true
			}
		}
	}
	return false
}