package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ranking struct {
	Id            string    `gorm:"type:uuid;primaryKey;column:id"`
	UserId        string    `gorm:"type:uuid;index;column:user_id"`
	QuizPackageId string    `gorm:"type:uuid;index;column:quiz_package_id"`
	TotalPoints   float64   `gorm:"type:float;default:0;column:total_points"`
	LastUpdated   time.Time `gorm:"type:timestamp;column:last_updated"`
	CreatedAt     time.Time `gorm:"type:timestamp;autoCreateTime;column:created_at"`

	User *User `json:"user" gorm:"foreignKey:UserId"`
}

func (Ranking) AllowedFields() map[string]bool {
	return map[string]bool{
		"user_id":         true,
		"quiz_package_id": true,
		"total_points":    true,
		"last_updated":    true,
	}
}

func (b *Ranking) BeforeCreate(tx *gorm.DB) (err error) {
	if b.Id == "" {
		b.Id = uuid.NewString()
	}
	return
}

func (b *Ranking) BeforeUpdate(tx *gorm.DB) (err error) {
	b.LastUpdated = time.Now()
	return
}
