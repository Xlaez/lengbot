package src

import "strings"

func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
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
