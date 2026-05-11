package controllers

import (
	"encoding/json"
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
		return utils.InternalServerError(ctx, errPkg.Error(), errPkg)
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
		Base: models.Base{
			Id:        utils.GenerateId(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserId:        user.Id,
		QuizPackageId: packageId,
		TransactionId: "",
		PurchasedDate: time.Now(),
		IsActive:      true,
		ExpiredDate:   time.Now().AddDate(0, 1, 0),
	}
	fmt.Println(utils.ToJson(&payment))
	errCreate := facades.Orm().Query().Create(&payment)
	if errCreate != nil {
		return utils.InternalServerError(ctx, errCreate.Error(), errCreate)
	}

	return utils.Ok(ctx, "Berhasil mendapatkan akses paket gratis", nil)
}

// ─── InitiatePaid ─────────────────────────────────────────────────────────────

// InitiatePaid membuat transaksi Midtrans Snap untuk paket berbayar.
// Flow:
//  1. Validasi request & cek ketersediaan paket
//  2. Buat Snap transaction di Midtrans → dapat snap_token & redirect_url
//  3. Simpan Transaction ke DB (status: pending)
//  4. Simpan UserPurchasedPackage ke DB (is_active: false, transaction_id → ID transaksi)
//  5. Return snap_token, redirect_url, dan order_id ke frontend
func (r *PaymentController) InitiatePaid(ctx http.Context) http.Response {
	data, err := utils.ValidateRequest(ctx, map[string]string{
		"packageId": "required|min_len:1",
		"method":    "required|in:gopay,bca_va,mandiri,bri_va,bni_va",
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

	// Ambil paket — tidak filter is_free karena ini untuk paket berbayar
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

	// Cek apakah user sudah memiliki paket ini (aktif)
	existingPurchase, _ := r.packageService.GetUserPurchasedPackage(user.Id, packageId)
	if existingPurchase != nil && existingPurchase.Id != "" && existingPurchase.IsActive {
		return utils.BadRequest(ctx, "Anda sudah memiliki paket ini", nil)
	}

	// Generate order ID unik yang akan dikirim ke Midtrans
	orderId := fmt.Sprintf("ORDER-%s-%s", packageId[:8], utils.GenerateId()[:8])

	// Map metode pembayaran user ke kode enabled_payments Midtrans Snap.
	//
	// Kode Snap untuk Virtual Account (langsung tersedia tanpa aktivasi khusus):
	//   bca_va   → Virtual Account BCA
	//   echannel → Mandiri Bill (kode khusus Mandiri di Midtrans)
	//   bri_va   → Virtual Account BRI
	//   bni_va   → Virtual Account BNI
	//
	// Kode Snap untuk e-wallet (perlu aktivasi di Midtrans Dashboard):
	//   gopay    → GoPay
	var enabledPayments []string
	var midtransPaymentMethod string
	switch method {
	case "gopay":
		enabledPayments = []string{"gopay"}
		midtransPaymentMethod = "gopay"
	case "bca_va":
		enabledPayments = []string{"bca_va"}
		midtransPaymentMethod = "bank_transfer"
	case "mandiri":
		enabledPayments = []string{"echannel"}
		midtransPaymentMethod = "echannel"
	case "bri_va":
		enabledPayments = []string{"bri_va"}
		midtransPaymentMethod = "bank_transfer"
	case "bni_va":
		enabledPayments = []string{"bni_va"}
		midtransPaymentMethod = "bank_transfer"
	default:
		enabledPayments = nil
		midtransPaymentMethod = "unknown"
	}

	// Panggil Midtrans Snap API untuk membuat transaksi
	snapResp, errSnap := r.midtransService.CreateSnapTransaction(services.MidtransSnapRequest{
		OrderId:         orderId,
		Amount:          int64(quizPackage.Price),
		PackageId:       packageId,
		PackageTitle:    quizPackage.Title,
		UserName:        user.Name,
		UserEmail:       user.Email,
		EnabledPayments: enabledPayments,
		UserPhone:       user.Phone,
	})
	if errSnap != nil {
		return utils.InternalServerError(ctx, "Gagal membuat transaksi pembayaran", errSnap.Error())
	}

	// ── Simpan transaksi ke DB (status: pending) ──────────────────────────────
	transaction := models.Transaction{
		UserId:          user.Id,
		QuizPackageId:   packageId,
		OrderId:         orderId,
		Amount:          quizPackage.Price,
		Currency:        quizPackage.Currency,
		PaymentMethod:   midtransPaymentMethod,
		PaymentProvider: "midtrans",
		Status:          "pending",
		SnapToken:       snapResp.Token,
		PaymentUrl:      snapResp.RedirectURL,
	}

	errCreateTx := facades.Orm().Query().Create(&transaction)
	if errCreateTx != nil {
		return utils.InternalServerError(ctx, "Gagal menyimpan transaksi", errCreateTx)
	}

	// ── Simpan UserPurchasedPackage sebagai pending (is_active: false) ────────
	// transaction_id di sini merujuk ke transactions.id (UUID) bukan order_id
	purchase := models.UserPurchasedPackage{
		UserId:        user.Id,
		QuizPackageId: packageId,
		TransactionId: transaction.Id, // FK ke transactions.id
		PurchasedDate: time.Now(),
		IsActive:      false, // akan di-set true setelah payment dikonfirmasi
		ExpiredDate:   time.Now().AddDate(0, 1, 0),
	}

	errCreatePurchase := facades.Orm().Query().Create(&purchase)
	if errCreatePurchase != nil {
		return utils.InternalServerError(ctx, "Gagal menyimpan data pembelian", errCreatePurchase)
	}

	return utils.Ok(ctx, "Transaksi berhasil dibuat", map[string]any{
		"snapToken":   snapResp.Token,
		"redirectUrl": snapResp.RedirectURL,
		"orderId":     orderId,
	})
}

// ─── GetPending ──────────────────────────────────────────────────────────────

// GetPending mengecek apakah user punya transaksi pending untuk paket tertentu.
// Dipanggil saat PaymentFlowScreen pertama kali dibuka.
func (r *PaymentController) GetPending(ctx http.Context) http.Response {
	packageId := ctx.Request().Route("package_id")
	if packageId == "" {
		return utils.BadRequest(ctx, "Package ID tidak ditemukan", nil)
	}

	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	var transaction models.Transaction
	facades.Orm().Query().
		Where("user_id", user.Id).
		Where("quiz_package_id", packageId).
		Where("status", "pending").
		Order("created_at DESC").
		First(&transaction)

	if transaction.Id == "" {
		return utils.Ok(ctx, "Tidak ada transaksi pending", map[string]any{
			"hasPending": false,
		})
	}

	return utils.Ok(ctx, "Ada transaksi pending", map[string]any{
		"hasPending": true,
		"transaction": map[string]any{
			"orderId":       transaction.OrderId,
			"snapToken":     transaction.SnapToken,
			"redirectUrl":   transaction.PaymentUrl,
			"amount":        transaction.Amount,
			"currency":      transaction.Currency,
			"paymentMethod": transaction.PaymentMethod,
			"createdAt":     transaction.CreatedAt,
		},
	})
}

// ─── CancelPendingTransaction ─────────────────────────────────────────────────

// CancelPendingTransaction membatalkan transaksi pending milik user.
// Dipanggil ketika user memilih "Ganti Metode Pembayaran".
func (r *PaymentController) CancelPendingTransaction(ctx http.Context) http.Response {
	orderId := ctx.Request().Route("order_id")
	if orderId == "" {
		return utils.BadRequest(ctx, "Order ID tidak ditemukan", nil)
	}

	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	// Pastikan transaksi ini milik user dan masih pending
	var transaction models.Transaction
	errFind := facades.Orm().Query().
		Where("order_id", orderId).
		Where("user_id", user.Id).
		Where("status", "pending").
		First(&transaction)

	if errFind != nil || transaction.Id == "" {
		return utils.NotFound(ctx, "Transaksi pending tidak ditemukan", nil)
	}

	// Batalkan di Midtrans (best effort — abaikan error jika sudah expire)
	r.midtransService.CancelTransaction(orderId) //nolint:errcheck

	// Update status di tabel transactions
	facades.Orm().Query().
		Where("order_id", orderId).
		Model(&models.Transaction{}).
		Update("status", "cancel")

	// Hapus record UserPurchasedPackage yang masih is_active=false untuk transaksi ini
	// agar tidak ada duplikat saat user membuat transaksi baru
	facades.Orm().Query().
		Where("transaction_id", transaction.Id).
		Where("is_active", false).
		Delete(&models.UserPurchasedPackage{})

	return utils.Ok(ctx, "Transaksi berhasil dibatalkan", nil)
}

// ─── Notification ─────────────────────────────────────────────────────────────

// Notification menerima webhook POST dari server Midtrans.
// Endpoint ini TIDAK memerlukan autentikasi (dipanggil oleh Midtrans server).
//
// Flow:
//  1. Parse body notifikasi
//  2. Verifikasi signature Midtrans
//  3. Cari Transaction berdasarkan order_id
//  4. Update status Transaction + simpan raw metadata dari Midtrans
//  5. Update is_active di UserPurchasedPackage sesuai status
func (r *PaymentController) Notification(ctx http.Context) http.Response {
	// Goravel otomatis parse JSON body via All()
	rawData := ctx.Request().All()

	// Ekstrak field dari notifikasi Midtrans
	orderId, _ := rawData["order_id"].(string)
	statusCode, _ := rawData["status_code"].(string)
	grossAmount, _ := rawData["gross_amount"].(string)
	signatureKey, _ := rawData["signature_key"].(string)
	transactionStatus, _ := rawData["transaction_status"].(string)
	midtransTransactionId, _ := rawData["transaction_id"].(string)
	paymentType, _ := rawData["payment_type"].(string)
	fraudStatus, _ := rawData["fraud_status"].(string)

	if orderId == "" {
		return utils.BadRequest(ctx, "Order ID tidak ditemukan dalam notifikasi", nil)
	}

	// Verifikasi bahwa notifikasi benar-benar dari Midtrans
	if !r.midtransService.VerifyNotificationSignature(orderId, statusCode, grossAmount, signatureKey) {
		return utils.BadRequest(ctx, "Signature tidak valid", nil)
	}

	// Cari transaksi berdasarkan order_id
	var transaction models.Transaction
	errFind := facades.Orm().Query().Where("order_id", orderId).First(&transaction)
	if errFind != nil || transaction.Id == "" {
		// Transaksi tidak ditemukan — tidak perlu error, cukup log dan return OK
		// (Midtrans kadang kirim notifikasi untuk transaksi yang sudah dihapus)
		return utils.Ok(ctx, "OK", nil)
	}

	// Serialize seluruh payload Midtrans sebagai metadata (untuk audit)
	metadataBytes, _ := json.Marshal(rawData)

	// Tentukan status final dan kapan pembayaran dikonfirmasi
	newStatus := transactionStatus
	var paidAt *time.Time

	switch transactionStatus {
	case "settlement", "capture":
		// Pembayaran berhasil dikonfirmasi
		if fraudStatus == "accept" || fraudStatus == "" {
			now := time.Now()
			paidAt = &now
			newStatus = "settlement"
		} else {
			// Terdeteksi fraud
			newStatus = "deny"
		}
	case "pending":
		newStatus = "pending"
	case "deny", "cancel", "expire", "failure":
		newStatus = transactionStatus
	}

	// Update record Transaction di DB
	updateMap := map[string]any{
		"status":                  newStatus,
		"midtrans_transaction_id": midtransTransactionId,
		"payment_method":          paymentType,
		"metadata":                string(metadataBytes),
		"updated_at":              time.Now(),
	}
	if paidAt != nil {
		updateMap["paid_at"] = paidAt
	}

	facades.Orm().Query().
		Where("order_id", orderId).
		Model(&models.Transaction{}).
		Update(updateMap)

	// Update UserPurchasedPackage berdasarkan status pembayaran
	switch newStatus {
	case "settlement":
		// Aktifkan akses paket
		facades.Orm().Query().
			Where("transaction_id", transaction.Id).
			Model(&models.UserPurchasedPackage{}).
			Update("is_active", true)

	case "deny", "cancel", "expire", "failure":
		// Non-aktifkan (jika sempat aktif sebelumnya)
		facades.Orm().Query().
			Where("transaction_id", transaction.Id).
			Model(&models.UserPurchasedPackage{}).
			Update("is_active", false)
	}

	return utils.Ok(ctx, "OK", nil)
}

// ─── CheckStatus ──────────────────────────────────────────────────────────────

// CheckStatus mengecek status pembayaran dari Midtrans dan menyinkronkan ke DB.
// Dipanggil oleh frontend setelah WebView selesai, sebagai fallback jika
// webhook Midtrans belum diterima.
//
// Flow:
//  1. Cari Transaction di DB berdasarkan order_id + user_id (verifikasi kepemilikan)
//  2. Panggil Midtrans Core API untuk status terkini
//  3. Update Transaction di DB jika status berubah
//  4. Update UserPurchasedPackage.is_active
//  5. Return status ke frontend
func (r *PaymentController) CheckStatus(ctx http.Context) http.Response {
	orderId := ctx.Request().Route("order_id")
	if orderId == "" {
		return utils.BadRequest(ctx, "Order ID tidak ditemukan", nil)
	}

	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	// Cari transaksi milik user ini berdasarkan order_id
	var transaction models.Transaction
	errFind := facades.Orm().Query().
		Where("order_id", orderId).
		Where("user_id", user.Id).
		First(&transaction)

	if errFind != nil || transaction.Id == "" {
		return utils.NotFound(ctx, "Transaksi tidak ditemukan", nil)
	}

	// Jika sudah settlement, tidak perlu cek ke Midtrans lagi
	if transaction.Status == "settlement" {
		return utils.Ok(ctx, "Pembayaran sudah dikonfirmasi", map[string]any{
			"orderId":           transaction.OrderId,
			"transactionStatus": transaction.Status,
			"paymentMethod":     transaction.PaymentMethod,
			"amount":            transaction.Amount,
			"currency":          transaction.Currency,
			"paidAt":            transaction.PaidAt,
		})
	}

	// Tanya Midtrans untuk status terkini
	status, errStatus := r.midtransService.CheckTransactionStatus(orderId)
	if errStatus != nil {
		return utils.InternalServerError(ctx, "Gagal mengecek status pembayaran ke Midtrans", errStatus.Error())
	}

	// Tentukan apakah status berubah dan perlu update DB
	if status.TransactionStatus != transaction.Status {
		var paidAt *time.Time
		newStatus := status.TransactionStatus

		if newStatus == "settlement" || newStatus == "capture" {
			now := time.Now()
			paidAt = &now
			newStatus = "settlement"
		}

		updateMap := map[string]any{
			"status":                  newStatus,
			"midtrans_transaction_id": status.TransactionId,
			"payment_method":          status.PaymentType,
			"updated_at":              time.Now(),
		}
		if paidAt != nil {
			updateMap["paid_at"] = paidAt
		}

		facades.Orm().Query().
			Where("order_id", orderId).
			Model(&models.Transaction{}).
			Update(updateMap)

		// Sinkronisasi UserPurchasedPackage
		switch newStatus {
		case "settlement":
			facades.Orm().Query().
				Where("transaction_id", transaction.Id).
				Model(&models.UserPurchasedPackage{}).
				Update("is_active", true)
		case "deny", "cancel", "expire", "failure":
			facades.Orm().Query().
				Where("transaction_id", transaction.Id).
				Model(&models.UserPurchasedPackage{}).
				Update("is_active", false)
		}

		transaction.Status = newStatus
		transaction.PaymentMethod = status.PaymentType
	}

	return utils.Ok(ctx, "Status pembayaran berhasil diambil", map[string]any{
		"orderId":           transaction.OrderId,
		"transactionStatus": transaction.Status,
		"paymentMethod":     transaction.PaymentMethod,
		"amount":            transaction.Amount,
		"currency":          transaction.Currency,
		"paidAt":            transaction.PaidAt,
	})
}
