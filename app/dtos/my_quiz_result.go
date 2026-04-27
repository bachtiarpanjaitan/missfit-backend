package dtos

type MyQuizResult struct {
	QuizPackageId string
	AvgScore      float64
	BestScore     float64
	HighestScore  float64
	Completed     int
	Passed        int
	TotalAttempts int
}
