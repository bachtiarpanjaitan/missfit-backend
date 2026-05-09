package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support"

	"missfit/app/facades"
	"missfit/app/http/controllers"
	"missfit/app/http/middleware"
	"missfit/app/services"
)

func Web() {
	// halaman default (biarin aja kalau masih butuh)
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("welcome.tmpl", map[string]any{
			"version": support.Version,
		})
	})

	// static file
	facades.Route().Static("public", "./public")

	// ─── Services ────────────────────────────────────────────────────────────
	packageService := services.NewPackageService()

	// Baca konfigurasi Midtrans dari environment
	midtransServerKey := facades.Config().Env("MIDTRANS_SERVER_KEY", "").(string)
	midtransEnv := facades.Config().Env("MIDTRANS_ENV", "sandbox").(string)
	midtransService := services.NewMidtransService(midtransServerKey, midtransEnv)

	// ─── Controllers ─────────────────────────────────────────────────────────
	authController := controllers.NewAuthController(packageService)
	quizController := controllers.NewQuizController(packageService)
	paymentController := controllers.NewPaymentController(packageService, midtransService)
	rankingController := controllers.NewRankingController(packageService)

	api := facades.Route().Prefix("/api")

	// ─── AUTH ─────────────────────────────────────────────────────────────────
	api.Post("/auth/register", authController.Register)
	api.Post("/auth/login", authController.Login)
	api.Middleware(middleware.Auth()).Get("/auth/me", authController.Me)

	// ─── QUIZZES ──────────────────────────────────────────────────────────────
	api.Middleware(middleware.Auth()).Get("/quizzes", quizController.Index)
	api.Middleware(middleware.Auth()).Get("/quizzes/all", quizController.All)
	api.Middleware(middleware.Auth()).Get("/quizzes/my-packages", quizController.MyPackages)
	api.Middleware(middleware.Auth()).Get("/quizzes/:package_id/questions", quizController.GetQuestions)
	api.Middleware(middleware.Auth()).Post("/quizzes/submit-result", quizController.SubmitResults)
	api.Middleware(middleware.Auth()).Get("/quizzes/my-quiz-stats", quizController.MyStats)

	// ─── PAYMENT ──────────────────────────────────────────────────────────────
	// Paket gratis — tidak butuh Midtrans
	api.Middleware(middleware.Auth()).Post("/payments/initiate-free", paymentController.InitiateFree)

	// Paket berbayar — buat Snap token Midtrans
	api.Middleware(middleware.Auth()).Post("/payments/initiate-paid", paymentController.InitiatePaid)

	// Webhook dari Midtrans — TIDAK pakai Auth middleware (dipanggil server Midtrans)
	api.Post("/payments/notification", paymentController.Notification)

	// Cek status transaksi manual oleh user
	api.Middleware(middleware.Auth()).Get("/payments/status/:order_id", paymentController.CheckStatus)

	// Cek transaksi pending untuk paket tertentu (dipanggil saat PaymentFlowScreen dibuka)
	api.Middleware(middleware.Auth()).Get("/payments/pending/:package_id", paymentController.GetPending)

	// Batalkan transaksi pending (saat user mau ganti metode pembayaran)
	api.Middleware(middleware.Auth()).Post("/payments/cancel-transaction/:order_id", paymentController.CancelPendingTransaction)

	// ─── RANKINGS ─────────────────────────────────────────────────────────────
	api.Middleware(middleware.Auth()).Get("/rankings/global", rankingController.GlobalRankings)
	api.Middleware(middleware.Auth()).Get("/rankings/package/:package_id", rankingController.PackageRank)
	api.Middleware(middleware.Auth()).Get("/rankings/my-rank", rankingController.MyRank)
}
