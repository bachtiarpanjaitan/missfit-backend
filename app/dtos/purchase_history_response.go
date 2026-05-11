package dtos

import "time"

// PurchaseHistoryItem merepresentasikan satu item riwayat pembelian paket.
// Data diambil dari tabel transactions JOIN quiz_packages.
type PurchaseHistoryItem struct {
	TransactionId string     `json:"transactionId"`
	OrderId       string     `json:"orderId"`
	PackageId     string     `json:"packageId"`
	PackageTitle  string     `json:"packageTitle"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	PaymentMethod string     `json:"paymentMethod"`
	Status        string     `json:"status"`
	PurchasedDate time.Time  `json:"purchasedDate"`
	PaidAt        *time.Time `json:"paidAt"`
}
