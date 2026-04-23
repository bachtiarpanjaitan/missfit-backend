package dtos

type QuizResult struct {
	UserId         string             `json:"userId"`
	PackageId      string             `json:"packageId"`
	Score          float64            `json:"score"`
	TotalQuestions float64            `json:"totalQuestions"`
	TimeSpent      float64            `json:"timeSpent"`
	StartedAt      string             `json:"startedAt"`
	CompletedAt    string             `json:"completedAt"`
	Answers        []QuizResultAnswer `json:"answers"`
}

type QuizResultAnswer struct {
	QuestionId string `json:"questionId"`
	AnswerId   string `json:"answerId"`
}
