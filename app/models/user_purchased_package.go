package models

import (
	"time"
)

type UserPurchasedPackage struct {
	Base

	UserId        string
	QuizPackageId string
	TransactionId string
	PurchasedDate time.Time
	IsActive      bool
	ExpiredDate   time.Time

	QuizPackage *QuizPackage `gorm:"foreignKey:QuizPackageId"`
}

func (UserPurchasedPackage) AllowedFields() map[string]bool {
	return map[string]bool{
		"user_id":         true,
		"quiz_package_id": true,
		"transaction_id":  true,
		"purchased_date":  true,
		"is_active":       true,
		"expired_date":    true,
	}
}
