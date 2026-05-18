package seeders

import (
	"fmt"
	"lumos/app/facades"
	"lumos/app/models"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *UserSeeder) Signature() string {
	return "UserSeeder"
}

// Run executes the seeder logic.
func (s *UserSeeder) Run() error {
	_, err := facades.Orm().Query().Exec("DELETE FROM users")
	if err != nil {
		return err
	}
	userPasswords := []string{"User1234", "User1234"}

	var users []models.User
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Admin1234"), bcrypt.DefaultCost)
	users = append(users, models.User{
		Base: models.Base{
			Id:        uuid.NewString(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:       "Admin",
		Username:   "admin",
		Role:       "admin",
		Email:      "admin@gmail.com",
		Password:   string(hashed),
		IsActive:   true,
		IsVerified: true,
	})

	for i, p := range userPasswords {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)

		users = append(users, models.User{
			Base: models.Base{
				Id:        uuid.NewString(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:       fmt.Sprintf("User %d", i+1),
			Username:   fmt.Sprintf("user%d", i+1),
			Role:       "user",
			Email:      fmt.Sprintf("user%d@gmail.com", i+1),
			Password:   string(hashed),
			IsActive:   true,
			IsVerified: true,
		})
	}

	return facades.Orm().Query().Create(&users)
}
