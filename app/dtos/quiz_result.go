package dtos

type QuizResult struct {
	UserId         string             `uuid:"userId"`
	PackageId      string             `uuid:"packageId"`
	Score          float64            `float:"score"`
	TotalQuestions float64            `int:"totalQuestions"`
	TimeSpent      float64            `int:"timeSpent"`
	StartedAt      string             `timestamp:"startedAt"`
	CompletedAt    string             `timestamp:"completedAt"`
	Answers        []QuizResultAnswer `json:"answers"`
}

type QuizResultAnswer struct {
	QuestionId string `uuid:"questionId"`
	AnswerId   string `uuid:"answerId"`
}
