package models

type UserQuizAnswer struct {
	Base

	UserQuizAttemptId string  `json:"user_quiz_attempt_id" gorm:"index"`
	SelectedOptionId  string  `json:"selected_option_id"`
	IsCorrect         bool    `json:"is_correct"`
	PointsEarned      float64 `json:"points_earned" gorm:"type:decimal(10,2)"`

	UserQuizAttempt *UserQuizAttempt `gorm:"foreignKey:UserQuizAttemptId"`
}
