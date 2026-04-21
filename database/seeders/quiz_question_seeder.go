package seeders

import (
	"time"

	"github.com/goravel/framework/facades"
)

type QuizQuestionSeeder struct{}

func (s *QuizQuestionSeeder) Signature() string {
	return "QuizQuestionSeeder"
}

func (s *QuizQuestionSeeder) Run() error {
	now := time.Now()

	data := []map[string]any{
		// PACKAGE 1
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440001",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440001",
			"question_text":      "Which macronutrient builds muscle?",
			"question_image_url": "https://api.example.com/questions/protein.jpg",
			"question_order":     1,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "Proteins build muscle",
			"point":              10,
			"created_at":         now,
		},
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440002",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440001",
			"question_text":      "Recommended calorie intake?",
			"question_image_url": nil,
			"question_order":     2,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "2000-2500 kcal",
			"point":              10,
			"created_at":         now,
		},

		// PACKAGE 2
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440003",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440002",
			"question_text":      "Exercise per day?",
			"question_image_url": nil,
			"question_order":     1,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "30 minutes",
			"point":              10,
			"created_at":         now,
		},
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440004",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440002",
			"question_text":      "Best exercise for heart?",
			"question_image_url": nil,
			"question_order":     2,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "Cardio",
			"point":              10,
			"created_at":         now,
		},

		// PACKAGE 3
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440005",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440003",
			"question_text":      "Vitamin C source?",
			"question_image_url": nil,
			"question_order":     1,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "Orange",
			"point":              10,
			"created_at":         now,
		},
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440006",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440003",
			"question_text":      "Protein source?",
			"question_image_url": nil,
			"question_order":     2,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "Egg",
			"point":              10,
			"created_at":         now,
		},

		// PACKAGE 4
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440007",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440004",
			"question_text":      "Muscle growth requires?",
			"question_image_url": nil,
			"question_order":     1,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "Training + protein",
			"point":              10,
			"created_at":         now,
		},
		{
			"id":                 "770e8400-e29b-41d4-a716-446655440008",
			"quiz_package_id":    "660e8400-e29b-41d4-a716-446655440004",
			"question_text":      "Cardio improves?",
			"question_image_url": nil,
			"question_order":     2,
			"question_type":      "multiple_choice",
			"number_of_options":  4,
			"explanation":        "Heart health",
			"point":              10,
			"created_at":         now,
		},
	}

	facades.DB().Table("quiz_questions").Insert(data)

	return nil
}
