package seeders

import (
	"time"

	"github.com/goravel/framework/facades"
)

type QuizPackageSeeder struct{}

func (s *QuizPackageSeeder) Signature() string {
	return "QuizPackageSeeder"
}

func (s *QuizPackageSeeder) Run() error {
	now := time.Now()

	data := []map[string]any{
		{
			"id":               "660e8400-e29b-41d4-a716-446655440001",
			"title":            "Basic Nutrition Quiz",
			"description":      "Learn about nutrition basics",
			"category":         "Nutrition",
			"difficulty_level": "easy",
			"thumbnail_url":    "https://images.unsplash.com/photo-1517694712202-14dd9538aa97?w=400&h=300&fit=crop",
			"is_free":          true,
			"price":            0,
			"currency":         "IDR",
			"total_questions":  5,
			"duration_minutes": 15,
			"passing_score":    60,
			"max_attempts":     5,
			"total_taken":      120,
			"average_score":    75.5,
			"is_published":     true,
			"published_at":     now,
			"created_at":       now,
		},
		{
			"id":               "660e8400-e29b-41d4-a716-446655440002",
			"title":            "Fitness Fundamentals",
			"description":      "Basic fitness knowledge",
			"category":         "Fitness",
			"difficulty_level": "easy",
			"thumbnail_url":    "https://picsum.photos/500/300",
			"is_free":          true,
			"price":            0,
			"currency":         "IDR",
			"total_questions":  5,
			"duration_minutes": 20,
			"passing_score":    65,
			"max_attempts":     3,
			"total_taken":      95,
			"average_score":    72.3,
			"is_published":     true,
			"published_at":     now,
			"created_at":       now,
		},
		{
			"id":               "660e8400-e29b-41d4-a716-446655440003",
			"title":            "Advanced Nutrition",
			"description":      "Deep nutrition concepts",
			"category":         "Nutrition",
			"difficulty_level": "medium",
			"thumbnail_url":    "https://fastly.picsum.photos/id/935/500/300.jpg",
			"is_free":          false,
			"price":            50000,
			"currency":         "IDR",
			"total_questions":  5,
			"duration_minutes": 25,
			"passing_score":    70,
			"max_attempts":     3,
			"total_taken":      40,
			"average_score":    68.5,
			"is_published":     true,
			"published_at":     now,
			"created_at":       now,
		},
		{
			"id":               "660e8400-e29b-41d4-a716-446655440004",
			"title":            "Workout Science",
			"description":      "Exercise physiology",
			"category":         "Fitness",
			"difficulty_level": "medium",
			"thumbnail_url":    "https://fastly.picsum.photos/id/341/500/300.jpg",
			"is_free":          false,
			"price":            60000,
			"currency":         "IDR",
			"total_questions":  5,
			"duration_minutes": 25,
			"passing_score":    70,
			"max_attempts":     3,
			"total_taken":      30,
			"average_score":    66.2,
			"is_published":     true,
			"published_at":     now,
			"created_at":       now,
		},
	}

	facades.DB().Table("quiz_packages").Insert(data)

	return nil
}
