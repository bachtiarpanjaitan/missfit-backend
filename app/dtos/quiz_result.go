package dtos

import "time"

type QuizResult struct {
	UserId         string             `uuid:"userId"`
	PackageId      string             `uuid:"packageId"`
	Score          float64            `float:"score"`
	TotalQuestions float64            `int:"totalQuestions"`
	TimeSpent      float64            `int:"timeSpent"`
	StartedAt      time.Time          `timestamp:"startedAt"`
	CompletedAt    time.Time          `timestamp:"completedAt"`
	Answers        []QuizResultAnswer `json:"answers"`
}

type QuizResultAnswer struct {
	QuestionId string `uuid:"questionId"`
	AnswerId   string `string:"answerId"`
}
