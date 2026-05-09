package controllers

import (
	"fmt"
	"time"

	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/services"
	"missfit/app/utils"

	"github.com/goravel/framework/contracts/http"
)

// ─── Controller struct ────────────────────────────────────────────────────────

type PaymentController struct {
	packageService  services.PackageServiceInterface
	midtransService services.MidtransServiceInterface
}

func NewPaymentController(
	packageService services.PackageServiceInterface,
	midtransService services.MidtransServiceInterface,
) *PaymentController {
	return &PaymentController{
		packageService:  packageService,
		midtransService: midtransService,
	}
}

// ─── InitiateFree ─────────────────────────────────────────────────────────────

// InitiateFree memberikan akses ke paket gratis tanpa pembayaran.
func (r *PaymentController) InitiateFree(ctx http.Context) http.Response {
	data, err := utils.ValidateRequest(ctx, map[string]string{
		"packageId": "required|min_len:1",
	})
	if err != nil {
		return err.(http.Response)
	}

	packageId := data["packageId"].(string)

	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	quizPackage, errPkg := r.packageService.GetPackageById(packageId, map[string]any{
		"is_free": true,
	})
	if errPkg != nil {
		return utils.InternalServerError(ctx, "Internal server error", errPkg)
	}

	if quizPackage == nil || quizPackage.Id == "" {
		return utils.BadRequest(ctx, "Paket Tidak Ditemukan", nil)
	}

	if !quizPackage.IsFree {
		return utils.BadRequest(ctx, "Paket Tidak Gratis", nil)
	}

	existingPurchase, _ := r.packageService.GetUserPurchasedPackage(user.Id, packageId)
	if existingPurchase != nil && existingPurchase.Id != "" {
		return utils.BadRequest(ctx, "Anda sudah membeli paket ini", nil)
	}

	payment := models.UserPurchasedPackage{
		UserId:        user.Id,
		QuizPackageId: packageId,
		TransactionId: "",
		PurchasedDate: time.Now(),
		IsActive:      true,
		ExpiredDate:   time.Now().AddDate(0, 1, 0),
	}

	errCreate := facades.Orm().Query().Create(&payment)
	if errCreate != nil {
		return utils.InternalServerError(ctx, "Internal server error", errCreate)
	}

	return utils.Ok(ctx, "Berhasil mendapatkan akses paket gratis", nil)
}

// ─── InitiatePaid ─────────────────────────────────────────────────────────────

// InitiatePaid membuat Midtrans Snap transaction untuk paket berbayar.
func (r *PaymentController) InitiatePaid(ctx http.Context) http.Response {
	data, err := utils.ValidateRequest(ctx, map[string]string{
		"packageId": "required|min_len:1",
		"method":    "required|in:dana,gopay,ovo,linkaja,card",
	})
	if err != nil {
		return err.(http.Response)
	}

	packageId := data["packageId"].(string)
	method := data["method"].(string)

	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	// Ambil paket (tanpa filter is_free — bisa paket berbayar)
	quizPackage, errPkg := r.packageService.GetPackageById(packageId, nil)
	if errPkg != nil {
		return utils.InternalServerError(ctx, "Internal server error", errPkg)
	}

	if quizPackage == nil || quizPackage.Id == "" {
		return utils.NotFound(ctx, "Paket tidak ditemukan", nil)
	}

	if quizPackage.IsFree {
		return utils.BadRequest(ctx, "Paket ini gratis, gunakan endpoint /payments/initiate-free", nil)
	}

	if !quizPackage.IsPublished {
		return utils.BadRequest(ctx, "Paket belum tersedia untuk dibeli", nil)
	}

	// Cek apakah user sudah membeli paket ini
	existingPurchase, _ := r.packageService.GetUserPurchasedPackage(user.Id, packageId)
	if existingPurchase != nil && existingPurchase.Id != "" {
		return utils.BadRequest(ctx, "Anda sudah membeli paket ini", nil)
	}

	// Generate order ID yang unik
	orderId := fmt.Sprintf("ORDER-%s-%s", packageId[:8], utils.GenerateId()[:8])

	// Map metode pembayaran ke Midtrans enabled_payments
	var enabledPayments []string
	switch method {
	case "gopay":
		enabledPayments = []string{"gopay"}
	case "dana":
		enabledPayments = []string{"qris"}
	case "ovo":
		enabledPayments = []string{"qris"}
	case "linkaja":
		enabledPayments = []string{"qris"}
	case "card":
		enabledPayments = []string{"credit_card"}
	default:
		enabledPayments = []string{"gopay"}
	}

	// Buat Snap transaction ke Midtrans
	snapResp, errSnap := r.midtransService.CreateSnapTransaction(services.MidtransSnapRequest{
		OrderId:         orderId,
		Amount:          int64(quizPackage.Price),
		PackageId:       packageId,
		PackageTitle:    quizPackage.Title,
		UserName:        user.Name,
		UserEmail:       user.Email,
		EnabledPayments: enabledPayments,
	})
	if errSnap != nil {
		return utils.InternalServerError(ctx, "Gagal membuat transaksi pembayaran", errSnap.Error())
	}

	// Simpan sebagai pending payment (IsActive = false sampai pembayaran dikonfirmasi)
	payment := models.UserPurchasedPackage{
		UserId:        user.Id,
		QuizPackageId: packageId,
		TransactionId: orderId,
		PurchasedDate: time.Now(),
		IsActive:      false,
		ExpiredDate:   time.Now().AddDate(0, 1, 0),
	}

	errCreate := facades.Orm().Query().Create(&payment)
	if errCreate != nil {
		return utils.InternalServerError(ctx, "Internal server error", errCreate)
	}

	return utils.Ok(ctx, "Berhasil membuat transaksi pembayaran", map[string]any{
		"snapToken":   snapResp.Token,
		"redirectUrl": snapResp.RedirectURL,
		"orderId":     orderId,
	})
}

// ─── Notification ─────────────────────────────────────────────────────────────

// Notification menerima dan memproses webhook notification dari Midtrans.
// Endpoint ini TIDAK memerlukan autentikasi (dipanggil oleh server Midtrans).
func (r *PaymentController) Notification(ctx http.Context) http.Response {
	// Goravel meng-parse JSON body secara otomatis via All()
	data := ctx.Request().All()

	// Ekstrak field dari notifikasi Midtrans
	orderId, _ := data["order_id"].(string)
	statusCode, _ := data["status_code"].(string)
	grossAmount, _ := data["gross_amount"].(string)
	signatureKey, _ := data["signature_key"].(string)
	transactionStatus, _ := data["transaction_status"].(string)
	transactionId, _ := data["transaction_id"].(string)
	paymentType, _ := data["payment_type"].(string)
	fraudStatus, _ := data["fraud_status"].(string)

	notification := services.MidtransTransactionStatus{
		OrderId:           orderId,
		TransactionId:     transactionId,
		TransactionStatus: transactionStatus,
		StatusCode:        statusCode,
		GrossAmount:       grossAmount,
		PaymentType:       paymentType,
		SignatureKey:      signatureKey,
		FraudStatus:       fraudStatus,
	}

	// Validasi bahwa order_id tidak kosong
	if notification.OrderId == "" {
		return utils.BadRequest(ctx, "Order ID tidak ditemukan dalam notifikasi", nil)
	}

	// Verifikasi signature untuk memastikan notifikasi berasal dari Midtrans
	if !r.midtransService.VerifyNotificationSignature(
		notification.OrderId,
		notification.StatusCode,
		notification.GrossAmount,
		notification.SignatureKey,
	) {
		return utils.BadRequest(ctx, "Signature tidak valid", nil)
	}

	// Proses berdasarkan status transaksi
	switch notification.TransactionStatus {
	case "settlement", "capture":
		// Pembayaran berhasil → aktifkan akses paket
		facades.Orm().Query().
			Where("transaction_id", notification.OrderId).
			Model(&models.UserPurchasedPackage{}).
			Update("is_active", true)

	case "deny", "cancel", "expire":
		// Pembayaran gagal/dibatalkan/kadaluarsa → non-aktifkan
		facades.Orm().Query().
			Where("transaction_id", notification.OrderId).
			Model(&models.UserPurchasedPackage{}).
			Update("is_active", false)

	case "pending":
		// Masih menunggu pembayaran — tidak perlu update
	}

	return utils.Ok(ctx, "OK", nil)
}

// ─── CheckStatus ──────────────────────────────────────────────────────────────

// CheckStatus mengecek status pembayaran dari Midtrans dan menyinkronkan ke DB.
func (r *PaymentController) CheckStatus(ctx http.Context) http.Response {
	orderId := ctx.Request().Route("order_id")
	if orderId == "" {
		return utils.BadRequest(ctx, "Order ID tidak ditemukan", nil)
	}

	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	// Cek kepemilikan: pastikan order ini milik user yang sedang login
	var purchase models.UserPurchasedPackage
	errFind := facades.Orm().Query().
		Where("transaction_id", orderId).
		Where("user_id", user.Id).
		First(&purchase)

	if errFind != nil || purchase.Id == "" {
		return utils.NotFound(ctx, "Transaksi tidak ditemukan", nil)
	}

	// Tanya Midtrans untuk status terkini
	status, errStatus := r.midtransService.CheckTransactionStatus(orderId)
	if errStatus != nil {
		return utils.InternalServerError(ctx, "Gagal mengecek status pembayaran", errStatus.Error())
	}

	// Sinkronisasi status ke DB
	switch status.TransactionStatus {
	case "settlement", "capture":
		facades.Orm().Query().
			Where("transaction_id", orderId).
			Model(&models.UserPurchasedPackage{}).
			Update("is_active", true)

	case "deny", "cancel", "expire":
		facades.Orm().Query().
			Where("transaction_id", orderId).
			Model(&models.UserPurchasedPackage{}).
			Update("is_active", false)
	}

	return utils.Ok(ctx, "Status pembayaran berhasil diambil", map[string]any{
		"orderId":           status.OrderId,
		"transactionId":     status.TransactionId,
		"transactionStatus": status.TransactionStatus,
		"paymentType":       status.PaymentType,
		"grossAmount":       status.GrossAmount,
		"fraudStatus":       status.FraudStatus,
	})
}
