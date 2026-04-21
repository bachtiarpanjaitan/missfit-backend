package seeders

import (
	"time"

	"github.com/goravel/framework/facades"
)

type QuizOptionSeeder struct{}

func (s *QuizOptionSeeder) Signature() string {
	return "QuizOptionSeeder"
}

func (s *QuizOptionSeeder) Run() error {
	now := time.Now()

	data := []map[string]any{
		// Q1
		{"id": "880e8400-e29b-41d4-a716-446655440001", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440001", "option_text": "Carbohydrates", "option_image_url": nil, "option_order": 1, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440002", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440001", "option_text": "Proteins", "option_image_url": nil, "option_order": 2, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440003", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440001", "option_text": "Fats", "option_image_url": nil, "option_order": 3, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440004", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440001", "option_text": "Fiber", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},

		// Q2
		{"id": "880e8400-e29b-41d4-a716-446655440005", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440002", "option_text": "1500-2000", "option_image_url": nil, "option_order": 1, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440006", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440002", "option_text": "2000-2500", "option_image_url": nil, "option_order": 2, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440007", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440002", "option_text": "3000+", "option_image_url": nil, "option_order": 3, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440008", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440002", "option_text": "4000+", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},

		// Q3
		{"id": "880e8400-e29b-41d4-a716-446655440009", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440003", "option_text": "10 min", "option_image_url": nil, "option_order": 1, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440010", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440003", "option_text": "20 min", "option_image_url": nil, "option_order": 2, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440011", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440003", "option_text": "30 min", "option_image_url": nil, "option_order": 3, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440012", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440003", "option_text": "60 min", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},

		// Q4
		{"id": "880e8400-e29b-41d4-a716-446655440013", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440004", "option_text": "Weightlifting", "option_image_url": nil, "option_order": 1, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440014", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440004", "option_text": "Cardio", "option_image_url": nil, "option_order": 2, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440015", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440004", "option_text": "Yoga", "option_image_url": nil, "option_order": 3, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440016", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440004", "option_text": "Stretching", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},

		// Q5
		{"id": "880e8400-e29b-41d4-a716-446655440017", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440005", "option_text": "Orange", "option_image_url": nil, "option_order": 1, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440018", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440005", "option_text": "Rice", "option_image_url": nil, "option_order": 2, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440019", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440005", "option_text": "Bread", "option_image_url": nil, "option_order": 3, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440020", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440005", "option_text": "Oil", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},

		// Q6
		{"id": "880e8400-e29b-41d4-a716-446655440021", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440006", "option_text": "Egg", "option_image_url": nil, "option_order": 1, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440022", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440006", "option_text": "Sugar", "option_image_url": nil, "option_order": 2, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440023", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440006", "option_text": "Salt", "option_image_url": nil, "option_order": 3, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440024", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440006", "option_text": "Water", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},

		// Q7
		{"id": "880e8400-e29b-41d4-a716-446655440025", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440007", "option_text": "Training + protein", "option_image_url": nil, "option_order": 1, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440026", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440007", "option_text": "Sleep only", "option_image_url": nil, "option_order": 2, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440027", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440007", "option_text": "Water only", "option_image_url": nil, "option_order": 3, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440028", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440007", "option_text": "Stretching", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},

		// Q8
		{"id": "880e8400-e29b-41d4-a716-446655440029", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440008", "option_text": "Heart health", "option_image_url": nil, "option_order": 1, "is_correct": true, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440030", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440008", "option_text": "Muscle only", "option_image_url": nil, "option_order": 2, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440031", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440008", "option_text": "Flexibility", "option_image_url": nil, "option_order": 3, "is_correct": false, "created_at": now},
		{"id": "880e8400-e29b-41d4-a716-446655440032", "quiz_question_id": "770e8400-e29b-41d4-a716-446655440008", "option_text": "Balance", "option_image_url": nil, "option_order": 4, "is_correct": false, "created_at": now},
	}

	facades.DB().Table("quiz_options").Insert(data)

	return nil
}
