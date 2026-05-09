package models

import (
	"encoding/json"
	"time"
)

// Transaction merekam setiap transaksi pembayaran via Midtrans.
// Satu Transaction terhubung ke satu UserPurchasedPackage melalui
// kolom transaction_id di tabel user_purchased_packages.
type Transaction struct {
	Base

	UserId        string
	QuizPackageId string

	// OrderId adalah order ID internal kita yang dikirim ke Midtrans
	// Format: "ORDER-{packageId[:8]}-{randomId[:8]}"
	OrderId string

	// MidtransTransactionId adalah ID yang di-generate oleh Midtrans
	// setelah pembayaran berhasil diproses
	MidtransTransactionId string

	Amount   float64
	Currency string

	// PaymentMethod adalah metode yang dipilih user (gopay, qris, credit_card, dll)
	PaymentMethod string

	// PaymentProvider selalu "midtrans" untuk saat ini
	PaymentProvider string

	// Status transaksi: pending | settlement | capture | deny | cancel | expire | failure
	Status string

	// SnapToken adalah token dari Midtrans Snap untuk membuka halaman pembayaran
	SnapToken string

	// PaymentUrl adalah URL redirect ke halaman Midtrans Snap
	PaymentUrl string

	// PaidAt terisi ketika status menjadi settlement atau capture
	PaidAt *time.Time

	// Metadata menyimpan raw response dari Midtrans notification sebagai JSON
	// berguna untuk audit dan debugging
	Metadata JSON

	// Relasi
	User        *User        `gorm:"foreignKey:UserId"`
	QuizPackage *QuizPackage `gorm:"foreignKey:QuizPackageId"`
}

// JSON adalah custom type untuk menyimpan JSONB di PostgreSQL.
type JSON json.RawMessage

func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return nil
	}
	*j = JSON(data)
	return nil
}

func (Transaction) AllowedFields() map[string]bool {
	return map[string]bool{
		"user_id":         true,
		"quiz_package_id": true,
		"order_id":        true,
		"status":          true,
		"amount":          true,
		"payment_method":  true,
		"created_at":      true,
	}
}
