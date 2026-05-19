package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ─── Interfaces ───────────────────────────────────────────────────────────────

type MidtransServiceInterface interface {
	CreateSnapTransaction(req MidtransSnapRequest) (*MidtransSnapResponse, error)
	CheckTransactionStatus(orderId string) (*MidtransTransactionStatus, error)
	CancelTransaction(orderId string) (*MidtransTransactionStatus, error)
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
	UserPhone       string
	EnabledPayments []string // e.g. ["gopay"], ["bca_va"], ["echannel"]

	// FinishCallbackUrl adalah URL yang digunakan WebView mobile untuk mendeteksi
	// bahwa pembayaran selesai. URL ini TIDAK perlu mengarah ke server sungguhan —
	// ia hanya berfungsi sebagai "sinyal" yang dicegat oleh onShouldStartLoadWithRequest
	// di React Native sebelum WebView sempat memuat URL tersebut.
	//
	// PERBEDAAN PENTING:
	//   FinishCallbackUrl (di sini)  → hanya untuk WebView interception, domain boleh fiktif
	//   Notification URL (Midtrans Dashboard) → harus server nyata, butuh ngrok untuk local dev
	//
	// Midtrans akan me-redirect WebView ke URL ini setelah pembayaran.
	// Boleh pakai pola apapun, contoh:
	//   "https://payment.ihandlumos.app/finish"  ← pola fiktif (default)
	//   "ihandlumos://payment/finish"            ← app deep link scheme
	FinishCallbackUrl   string
	UnfinishCallbackUrl string
	ErrorCallbackUrl    string
}

// MidtransSnapResponse adalah respons sukses dari Snap API.
type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

// MidtransErrorResponse adalah respons error dari Midtrans API.
// Digunakan untuk mengekstrak pesan error yang lebih informatif.
type MidtransErrorResponse struct {
	StatusCode         string   `json:"status_code"`
	StatusMessage      string   `json:"status_message"`
	ValidationMessages []string `json:"validation_messages"`
}

func (e *MidtransErrorResponse) Error() string {
	msg := fmt.Sprintf("Midtrans error [%s]: %s", e.StatusCode, e.StatusMessage)
	if len(e.ValidationMessages) > 0 {
		msg += fmt.Sprintf(" — validation: %v", e.ValidationMessages)
	}
	return msg
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

// URL default untuk callback WebView. Pola ini dicegat oleh onShouldStartLoadWithRequest
// di React Native — tidak perlu mengarah ke server nyata.
const (
	defaultFinishUrl   = "https://payment.ihandlumos.app/finish"
	defaultUnfinishUrl = "https://payment.ihandlumos.app/unfinish"
	defaultErrorUrl    = "https://payment.ihandlumos.app/error"
)

func (r MidtransSnapRequest) finishUrl() string {
	if r.FinishCallbackUrl != "" {
		return r.FinishCallbackUrl
	}
	return defaultFinishUrl
}

func (r MidtransSnapRequest) unfinishUrl() string {
	if r.UnfinishCallbackUrl != "" {
		return r.UnfinishCallbackUrl
	}
	return defaultUnfinishUrl
}

func (r MidtransSnapRequest) errorUrl() string {
	if r.ErrorCallbackUrl != "" {
		return r.ErrorCallbackUrl
	}
	return defaultErrorUrl
}

// ─── Service implementation ───────────────────────────────────────────────────

type MidtransService struct {
	serverKey  string
	env        string // "sandbox" | "production"
	httpClient *http.Client
}

// NewMidtransService membuat instance baru MidtransService.
func NewMidtransService(serverKey, env string) MidtransServiceInterface {
	return &MidtransService{
		serverKey: serverKey,
		env:       env,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getSnapURL mengembalikan endpoint Snap API sesuai environment.
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
// Format Midtrans: base64("serverKey:") — password dikosongkan (hanya username)
func (s *MidtransService) getAuthHeader() string {
	encoded := base64.StdEncoding.EncodeToString([]byte(s.serverKey + ":"))
	return "Basic " + encoded
}

// ─── CreateSnapTransaction ────────────────────────────────────────────────────

// CreateSnapTransaction membuat Snap payment token via Midtrans Snap API.
//
// PENTING: enabled_payments TIDAK boleh dikirim jika kosong/nil — Midtrans akan
// return error 400. Field ini hanya disertakan jika ada isinya.
func (s *MidtransService) CreateSnapTransaction(req MidtransSnapRequest) (*MidtransSnapResponse, error) {
	// Bangun payload dasar — field wajib selalu ada
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
			"phone":      req.UserPhone,
			"billing_address": map[string]any{
				"first_name":   req.UserName,
				"email":        req.UserEmail,
				"phone":        req.UserPhone,
				"country_code": "IDN",
			},
		},
		// Callback URL untuk redirect WebView setelah pembayaran.
		// Gunakan nilai default jika tidak diisi.
		"callbacks": map[string]any{
			"finish":   req.finishUrl(),
			"unfinish": req.unfinishUrl(),
			"error":    req.errorUrl(),
		},
	}

	// FIX: Hanya tambahkan enabled_payments jika ada isinya.
	// Mengirim array kosong [] atau null akan membuat Midtrans return 400.
	if len(req.EnabledPayments) > 0 {
		payload["enabled_payments"] = req.EnabledPayments
	}

	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("midtrans: gagal marshal payload: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.getSnapURL(), bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("midtrans: gagal membuat HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", s.getAuthHeader())

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans: HTTP request gagal: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("midtrans: gagal membaca response body: %w", err)
	}

	// Jika bukan 2xx, parse error response dari Midtrans untuk pesan yang jelas
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var midtransErr MidtransErrorResponse
		if jsonErr := json.Unmarshal(respBody, &midtransErr); jsonErr == nil && midtransErr.StatusMessage != "" {
			return nil, &midtransErr
		}
		// Fallback jika tidak bisa parse sebagai MidtransErrorResponse
		return nil, fmt.Errorf("midtrans: Snap API status %d — %s", resp.StatusCode, string(respBody))
	}

	var snapResp MidtransSnapResponse
	if err := json.Unmarshal(respBody, &snapResp); err != nil {
		return nil, fmt.Errorf("midtrans: gagal parse Snap response: %w", err)
	}

	if snapResp.Token == "" {
		return nil, fmt.Errorf("midtrans: snap token kosong, response: %s", string(respBody))
	}

	return &snapResp, nil
}

// ─── CheckTransactionStatus ──────────────────────────────────────────────────

// CheckTransactionStatus mengambil status transaksi dari Midtrans Core API.
func (s *MidtransService) CheckTransactionStatus(orderId string) (*MidtransTransactionStatus, error) {
	url := fmt.Sprintf("%s/%s/status", s.getAPIBaseURL(), orderId)

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("midtrans: gagal membuat HTTP request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", s.getAuthHeader())

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans: HTTP request gagal: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("midtrans: gagal membaca response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var midtransErr MidtransErrorResponse
		if jsonErr := json.Unmarshal(respBody, &midtransErr); jsonErr == nil && midtransErr.StatusMessage != "" {
			return nil, &midtransErr
		}
		return nil, fmt.Errorf("midtrans: Status API status %d — %s", resp.StatusCode, string(respBody))
	}

	var status MidtransTransactionStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("midtrans: gagal parse status response: %w", err)
	}

	return &status, nil
}

// ─── CancelTransaction ──────────────────────────────────────────────────────

// CancelTransaction membatalkan transaksi di Midtrans via Core API.
// Digunakan ketika user ingin mengganti metode pembayaran.
func (s *MidtransService) CancelTransaction(orderId string) (*MidtransTransactionStatus, error) {
	url := fmt.Sprintf("%s/%s/cancel", s.getAPIBaseURL(), orderId)

	httpReq, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("midtrans: gagal membuat HTTP request cancel: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", s.getAuthHeader())

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans: HTTP request cancel gagal: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("midtrans: gagal membaca response cancel: %w", err)
	}

	var status MidtransTransactionStatus
	// Abaikan error parse — transaksi mungkin sudah expire/tidak ada di Midtrans
	json.Unmarshal(respBody, &status) //nolint:errcheck

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
