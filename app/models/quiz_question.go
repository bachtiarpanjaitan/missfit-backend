package models

type QuizQuestion struct {
	Base

	QuizPackageId    string  `json:"quiz_package_id"`
	QuestionText     string  `json:"question_text"`
	QuestionImageUrl string  `json:"question_image_url"`
	QuestionOrder    int     `json:"question_order"`
	QuestionType     string  `json:"question_type"`
	NumberOfOptions  int     `json:"number_of_options"`
	Explanation      string  `json:"explanation"`
	Point            float64 `json:"point"`

	QuizPackage *QuizPackage `gorm:"foreignKey:QuizPackageId"`
	Options     []QuizOption `gorm:"foreignKey:QuizQuestionId"`
}

func (QuizQuestion) AllowedFields() map[string]bool {
	return map[string]bool{
		"quiz_package_id":    true,
		"question_text":      true,
		"question_image_url": true,
		"question_order":     true,
		"question_type":      true,
		"number_of_options":  true,
		"explaination":       true,
		"point":              true,
	}
}
