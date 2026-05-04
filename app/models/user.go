package models

import (
	"time"
)

type User struct {
	Base

	Role                  string
	Name                  string
	Email                 string `gorm:"unique"`
	Username              string `gorm:"unique"`
	Password              string `json:"-"`
	AvatarURL             string
	Bio                   *string
	DateOfBirth           *time.Time
	Gender                string
	Phone                 *string
	TotalPoints           float64
	TotalQuizzesCompleted int
	AuthProvider          string
	AuthProviderID        string
	IsVerified            bool
	IsActive              bool
	LastLoginAt           *time.Time
}

func (User) AllowedFields() map[string]bool {
	return map[string]bool{
		"role":  true,
		"email": true,
		"name":  true,
	}
}
