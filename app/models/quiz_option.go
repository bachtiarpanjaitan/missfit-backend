package models

type QuizOption struct {
	Base

	QuizQuestionId string `json:"quiz_question_id"`
	OptionText     string `json:"option_text"`
	OptionImageUrl string `json:"option_image_url"`
	OptionOrder    int    `json:"option_order"`
	IsCorrect      bool   `json:"is_correct"`

	QuizQuestion *QuizQuestion `gorm:"foreignKey:QuizQuestionId"`
}

func (QuizOption) AllowedFields() map[string]bool {
	return map[string]bool{
		"quiz_question_id": true,
		"option_text":      true,
		"option_image_url": true,
		"option_order":     true,
		"is_correct":       true,
	}
}
