package models

import (
	"time"
)

type QuizPackage struct {
	Base

	Title           string
	Description     string
	Category        string
	DifficultyLevel string
	ThumbnailUrl    string
	EducationLevel  string
	Price           float64
	IsFree          bool
	Currency        string
	TotalQuestions  int
	DurationMinutes int
	PassingScore    int
	MaxAttempts     int
	TotalTaken      int
	AverageScore    float64
	Rating          float64
	IsPublished     bool
	PublishedAt     *time.Time
}

func (QuizPackage) AllowedFields() map[string]bool {
	return map[string]bool{
		"title":           true,
		"description":     true,
		"category":        true,
		"education_level": true,
		"is_free":         true,
		"created_at":      true,
		"rating":          true,
	}
}

// func (QuizPackage) Questions() []QuizQuestion {
// 	return nil
// }
