package services

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testJSONResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func newTestMidtransService(serverKey, env string, fn roundTripFunc) *MidtransService {
	return &MidtransService{
		serverKey: serverKey,
		env:       env,
		httpClient: &http.Client{
			Transport: fn,
		},
	}
}

func TestMidtransSnapRequestCallbackURLs(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		req := MidtransSnapRequest{}

		assert.Equal(t, defaultFinishUrl, req.finishUrl())
		assert.Equal(t, defaultUnfinishUrl, req.unfinishUrl())
		assert.Equal(t, defaultErrorUrl, req.errorUrl())
	})

	t.Run("uses custom values", func(t *testing.T) {
		req := MidtransSnapRequest{
			FinishCallbackUrl:   "missfit://finish",
			UnfinishCallbackUrl: "missfit://unfinish",
			ErrorCallbackUrl:    "missfit://error",
		}

		assert.Equal(t, "missfit://finish", req.finishUrl())
		assert.Equal(t, "missfit://unfinish", req.unfinishUrl())
		assert.Equal(t, "missfit://error", req.errorUrl())
	})
}

func TestMidtransServiceURLsAndAuthHeader(t *testing.T) {
	sandbox := &MidtransService{serverKey: "server-key", env: "sandbox"}
	production := &MidtransService{serverKey: "server-key", env: "production"}

	assert.Equal(t, "https://app.sandbox.midtrans.com/snap/v1/transactions", sandbox.getSnapURL())
	assert.Equal(t, "https://api.sandbox.midtrans.com/v2", sandbox.getAPIBaseURL())
	assert.Equal(t, "https://app.midtrans.com/snap/v1/transactions", production.getSnapURL())
	assert.Equal(t, "https://api.midtrans.com/v2", production.getAPIBaseURL())
	assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("server-key:")), sandbox.getAuthHeader())
}

func TestMidtransErrorResponseError(t *testing.T) {
	errResp := MidtransErrorResponse{
		StatusCode:         "400",
		StatusMessage:      "validation failed",
		ValidationMessages: []string{"gross_amount is required"},
	}

	assert.Contains(t, errResp.Error(), "Midtrans error [400]: validation failed")
	assert.Contains(t, errResp.Error(), "gross_amount is required")
}

func TestMidtransServiceCreateSnapTransaction(t *testing.T) {
	t.Run("sends payload without empty enabled payments", func(t *testing.T) {
		service := newTestMidtransService("server-key", "sandbox", func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "https://app.sandbox.midtrans.com/snap/v1/transactions", req.URL.String())
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", req.Header.Get("Accept"))
			assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("server-key:")), req.Header.Get("Authorization"))

			var payload map[string]any
			require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
			assert.NotContains(t, payload, "enabled_payments")

			transactionDetails := payload["transaction_details"].(map[string]any)
			assert.Equal(t, "order-1", transactionDetails["order_id"])
			assert.Equal(t, float64(150000), transactionDetails["gross_amount"])

			callbacks := payload["callbacks"].(map[string]any)
			assert.Equal(t, defaultFinishUrl, callbacks["finish"])
			assert.Equal(t, defaultUnfinishUrl, callbacks["unfinish"])
			assert.Equal(t, defaultErrorUrl, callbacks["error"])

			return testJSONResponse(http.StatusCreated, `{"token":"snap-token","redirect_url":"https://snap.test"}`), nil
		})

		resp, err := service.CreateSnapTransaction(MidtransSnapRequest{
			OrderId:      "order-1",
			Amount:       150000,
			PackageId:    "pkg-1",
			PackageTitle: "Paket A",
			UserName:     "User A",
			UserEmail:    "user@example.com",
			UserPhone:    "08123",
		})

		require.NoError(t, err)
		assert.Equal(t, "snap-token", resp.Token)
		assert.Equal(t, "https://snap.test", resp.RedirectURL)
	})

	t.Run("sends enabled payments and custom callbacks", func(t *testing.T) {
		service := newTestMidtransService("server-key", "production", func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://app.midtrans.com/snap/v1/transactions", req.URL.String())

			var payload map[string]any
			require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))

			assert.Equal(t, []any{"gopay", "bca_va"}, payload["enabled_payments"])
			callbacks := payload["callbacks"].(map[string]any)
			assert.Equal(t, "missfit://finish", callbacks["finish"])
			assert.Equal(t, "missfit://unfinish", callbacks["unfinish"])
			assert.Equal(t, "missfit://error", callbacks["error"])

			return testJSONResponse(http.StatusOK, `{"token":"snap-token","redirect_url":"https://snap.test"}`), nil
		})

		resp, err := service.CreateSnapTransaction(MidtransSnapRequest{
			OrderId:             "order-2",
			Amount:              200000,
			PackageId:           "pkg-2",
			PackageTitle:        "Paket B",
			UserName:            "User B",
			UserEmail:           "user-b@example.com",
			UserPhone:           "08456",
			EnabledPayments:     []string{"gopay", "bca_va"},
			FinishCallbackUrl:   "missfit://finish",
			UnfinishCallbackUrl: "missfit://unfinish",
			ErrorCallbackUrl:    "missfit://error",
		})

		require.NoError(t, err)
		assert.Equal(t, "snap-token", resp.Token)
	})

	t.Run("returns Midtrans error response", func(t *testing.T) {
		service := newTestMidtransService("server-key", "sandbox", func(req *http.Request) (*http.Response, error) {
			return testJSONResponse(http.StatusBadRequest, `{
				"status_code":"400",
				"status_message":"validation failed",
				"validation_messages":["gross_amount is required"]
			}`), nil
		})

		resp, err := service.CreateSnapTransaction(MidtransSnapRequest{})

		require.Nil(t, resp)
		var midtransErr *MidtransErrorResponse
		require.ErrorAs(t, err, &midtransErr)
		assert.Equal(t, "400", midtransErr.StatusCode)
		assert.Equal(t, "validation failed", midtransErr.StatusMessage)
	})

	t.Run("rejects success response with empty token", func(t *testing.T) {
		service := newTestMidtransService("server-key", "sandbox", func(req *http.Request) (*http.Response, error) {
			return testJSONResponse(http.StatusOK, `{"redirect_url":"https://snap.test"}`), nil
		})

		resp, err := service.CreateSnapTransaction(MidtransSnapRequest{})

		require.Nil(t, resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "snap token kosong")
	})
}

func TestMidtransServiceCheckTransactionStatus(t *testing.T) {
	t.Run("returns parsed status", func(t *testing.T) {
		service := newTestMidtransService("server-key", "sandbox", func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "https://api.sandbox.midtrans.com/v2/order-1/status", req.URL.String())
			assert.Equal(t, "application/json", req.Header.Get("Accept"))
			assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("server-key:")), req.Header.Get("Authorization"))

			return testJSONResponse(http.StatusOK, `{
				"order_id":"order-1",
				"transaction_id":"trx-1",
				"transaction_status":"settlement",
				"status_code":"200",
				"gross_amount":"150000.00",
				"payment_type":"gopay"
			}`), nil
		})

		status, err := service.CheckTransactionStatus("order-1")

		require.NoError(t, err)
		assert.Equal(t, "order-1", status.OrderId)
		assert.Equal(t, "settlement", status.TransactionStatus)
		assert.Equal(t, "gopay", status.PaymentType)
	})

	t.Run("returns Midtrans error response", func(t *testing.T) {
		service := newTestMidtransService("server-key", "sandbox", func(req *http.Request) (*http.Response, error) {
			return testJSONResponse(http.StatusNotFound, `{"status_code":"404","status_message":"not found"}`), nil
		})

		status, err := service.CheckTransactionStatus("missing-order")

		require.Nil(t, status)
		var midtransErr *MidtransErrorResponse
		require.ErrorAs(t, err, &midtransErr)
		assert.Equal(t, "404", midtransErr.StatusCode)
	})
}

func TestMidtransServiceCancelTransaction(t *testing.T) {
	service := newTestMidtransService("server-key", "production", func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "https://api.midtrans.com/v2/order-1/cancel", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Accept"))
		assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("server-key:")), req.Header.Get("Authorization"))

		return testJSONResponse(http.StatusNotFound, `{
			"order_id":"order-1",
			"transaction_status":"cancel",
			"status_code":"404"
		}`), nil
	})

	status, err := service.CancelTransaction("order-1")

	require.NoError(t, err)
	assert.Equal(t, "order-1", status.OrderId)
	assert.Equal(t, "cancel", status.TransactionStatus)
	assert.Equal(t, "404", status.StatusCode)
}

func TestMidtransServiceVerifyNotificationSignature(t *testing.T) {
	service := &MidtransService{serverKey: "server-key"}
	raw := "order-1" + "200" + "150000.00" + "server-key"
	hash := sha512.Sum512([]byte(raw))
	signature := fmt.Sprintf("%x", hash)

	assert.True(t, service.VerifyNotificationSignature("order-1", "200", "150000.00", signature))
	assert.False(t, service.VerifyNotificationSignature("order-1", "200", "150000.00", "invalid-signature"))
}
