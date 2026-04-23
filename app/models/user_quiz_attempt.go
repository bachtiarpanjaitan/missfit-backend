package models

import "time"

type UserQuizAttempt struct {
	Base

	UserId           string    `json:"user_id" gorm:"index"`
	QuizPackageId    string    `json:"quiz_package_id" gorm:"index"`
	StartedAt        time.Time `json:"started_at" gorm:"type:timestamptz"`
	CompletedAt      time.Time `json:"completed_at" gorm:"type:timestamptz"`
	Score            float64   `json:"score" gorm:"type:decimal(10,2)"`
	TotalPoints      float64   `json:"total_points" gorm:"type:decimal(10,2)"`
	IsPassed         *bool     `json:"is_passed"`
	Percentage       float64   `json:"percentage" gorm:"type:decimal(10,2)"`
	Status           string    `json:"status" gorm:"index"`
	TimeTakenSeconds int64     `json:"time_taken_seconds" gorm:"type:bigint"`

	User            *User            `json:"user" gorm:"foreignKey:UserId"`
	QuizPackage     *QuizPackage     `json:"quiz_package" gorm:"foreignKey:QuizPackageId"`
	UserQuizAnswers []UserQuizAnswer `json:"user_quiz_answers" gorm:"foreignKey:UserQuizAttemptId"`
}
