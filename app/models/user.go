package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    string `gorm:"primaryKey"`
	Name                  string
	Email                 string `gorm:"unique"`
	Username              string `gorm:"unique"`
	Password              string `json:"-"`
	FullName              string
	AvatarURL             string
	Bio                   *string
	DateOfBirth           *time.Time
	Gender                string
	Phone                 *string
	TotalPoints           int
	TotalQuizzesCompleted int
	AuthProvider          string
	AuthProviderID        string
	IsVerified            bool
	IsActive              bool
	LastLoginAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

func (u *User) BeforeCreate() (err error) {
	u.ID = uuid.NewString()
	return
}
