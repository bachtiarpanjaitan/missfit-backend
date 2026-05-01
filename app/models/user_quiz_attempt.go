package models

import "time"

type UserQuizAttempt struct {
	Base

	UserId           string    `gorm:"column:user_id;index"`
	QuizPackageId    string    `gorm:"column:quiz_package_id;index"`
	StartedAt        time.Time `gorm:"column:started_at;type:timestamptz"`
	CompletedAt      time.Time `gorm:"column:completed_at;type:timestamptz"`
	Score            float64   `gorm:"column:score;type:decimal(10,2)"`
	TotalPoints      float64   `gorm:"column:total_points;type:decimal(10,2)"`
	IsPassed         *bool     `gorm:"column:is_passed"`
	Percentage       float64   `gorm:"column:percentage;type:decimal(10,2)"`
	Status           string    `gorm:"column:status;index"`
	TimeTakenSeconds int64     `gorm:"column:time_taken_seconds;type:bigint"`
	CorrectAnswers   int       `gorm:"column:correct_answers;type:int"`
	WrongAnswers     int       `gorm:"column:wrong_answers;type:int"`
	SkipAnswers      int       `gorm:"column:skip_answers;type:int"`

	User            *User            `json:"user" gorm:"foreignKey:UserId"`
	QuizPackage     *QuizPackage     `json:"quiz_package" gorm:"foreignKey:QuizPackageId"`
	UserQuizAnswers []UserQuizAnswer `json:"user_quiz_answers" gorm:"foreignKey:UserQuizAttemptId"`
}
