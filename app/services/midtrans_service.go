package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ─── Interfaces ───────────────────────────────────────────────────────────────

type MidtransServiceInterface interface {
	CreateSnapTransaction(req MidtransSnapRequest) (*MidtransSnapResponse, error)
	CheckTransactionStatus(orderId string) (*MidtransTransactionStatus, error)
	VerifyNotificationSignature(orderId, statusCode, grossAmount, signatureKey string) bool
}

// ─── Request / Response structs ───────────────────────────────────────────────

// MidtransSnapRequest berisi data yang dibutuhkan untuk membuat Snap transaction.
type MidtransSnapRequest struct {
	OrderId         string
	Amount          int64
	PackageId       string
	PackageTitle    string
	UserName        string
	UserEmail       string
	EnabledPayments []string // e.g. ["gopay"], ["qris"], ["credit_card"]
}

// MidtransSnapResponse adalah respons dari Snap API.
type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

// MidtransTransactionStatus adalah respons dari Status API.
type MidtransTransactionStatus struct {
	OrderId           string `json:"order_id"`
	TransactionId     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	SignatureKey      string `json:"signature_key"`
	FraudStatus       string `json:"fraud_status"`
}

// ─── Service implementation ───────────────────────────────────────────────────

type MidtransService struct {
	serverKey string
	env       string // "sandbox" | "production"
}

// NewMidtransService membuat instance baru MidtransService menggunakan
// constructor injection (tidak langsung ke facades agar mudah di-test).
func NewMidtransService(serverKey, env string) MidtransServiceInterface {
	return &MidtransService{
		serverKey: serverKey,
		env:       env,
	}
}

// getSnapURL mengembalikan base URL Snap API sesuai environment.
func (s *MidtransService) getSnapURL() string {
	if s.env == "production" {
		return "https://app.midtrans.com/snap/v1/transactions"
	}
	return "https://app.sandbox.midtrans.com/snap/v1/transactions"
}

// getAPIBaseURL mengembalikan base URL Core API sesuai environment.
func (s *MidtransService) getAPIBaseURL() string {
	if s.env == "production" {
		return "https://api.midtrans.com/v2"
	}
	return "https://api.sandbox.midtrans.com/v2"
}

// getAuthHeader menghasilkan Basic Auth header dari server key.
// Format Midtrans: base64("serverKey:")
func (s *MidtransService) getAuthHeader() string {
	encoded := base64.StdEncoding.EncodeToString([]byte(s.serverKey + ":"))
	return "Basic " + encoded
}

// ─── CreateSnapTransaction ────────────────────────────────────────────────────

// CreateSnapTransaction membuat Snap payment token via Midtrans Snap API.
func (s *MidtransService) CreateSnapTransaction(req MidtransSnapRequest) (*MidtransSnapResponse, error) {
	payload := map[string]any{
		"transaction_details": map[string]any{
			"order_id":     req.OrderId,
			"gross_amount": req.Amount,
		},
		"item_details": []map[string]any{
			{
				"id":       req.PackageId,
				"price":    req.Amount,
				"quantity": 1,
				"name":     req.PackageTitle,
			},
		},
		"customer_details": map[string]any{
			"first_name": req.UserName,
			"email":      req.UserEmail,
		},
		"enabled_payments": req.EnabledPayments,
		"callbacks": map[string]any{
			"finish": "https://payment.missfit.app/finish",
		},
	}

	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("midtrans: failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.getSnapURL(), bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("midtrans: failed to build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", s.getAuthHeader())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("midtrans: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("midtrans: Snap API returned status %d — body: %s", resp.StatusCode, string(respBody))
	}

	var snapResp MidtransSnapResponse
	if err := json.Unmarshal(respBody, &snapResp); err != nil {
		return nil, fmt.Errorf("midtrans: failed to unmarshal Snap response: %w", err)
	}

	return &snapResp, nil
}

// ─── CheckTransactionStatus ──────────────────────────────────────────────────

// CheckTransactionStatus mengambil status transaksi dari Midtrans Core API.
func (s *MidtransService) CheckTransactionStatus(orderId string) (*MidtransTransactionStatus, error) {
	url := fmt.Sprintf("%s/%s/status", s.getAPIBaseURL(), orderId)

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("midtrans: failed to build HTTP request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", s.getAuthHeader())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("midtrans: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("midtrans: Status API returned status %d — body: %s", resp.StatusCode, string(respBody))
	}

	var status MidtransTransactionStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("midtrans: failed to unmarshal status response: %w", err)
	}

	return &status, nil
}

// ─── VerifyNotificationSignature ─────────────────────────────────────────────

// VerifyNotificationSignature memverifikasi signature dari Midtrans notification.
// Formula: SHA512(order_id + status_code + gross_amount + server_key)
func (s *MidtransService) VerifyNotificationSignature(orderId, statusCode, grossAmount, signatureKey string) bool {
	raw := orderId + statusCode + grossAmount + s.serverKey
	hash := sha512.Sum512([]byte(raw))
	computed := fmt.Sprintf("%x", hash)
	return computed == signatureKey
}
