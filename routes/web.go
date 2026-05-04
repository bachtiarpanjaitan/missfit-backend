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

	//services
	packageService := services.NewPackageService()

	// controllers
	// userController := controllers.NewUserController()
	authController := &controllers.AuthController{}
	quizController := controllers.NewQuizController(packageService)
	paymentController := controllers.NewPaymentController(packageService)
	rankingController := controllers.NewRankingController(packageService)

	api := facades.Route().Prefix("/api")

	//AUTH
	api.Post("/auth/register", authController.Register)
	api.Post("/auth/login", authController.Login)
	api.Middleware(middleware.Auth()).Get("/auth/me", authController.Me)

	//QUIZZES
	api.Middleware(middleware.Auth()).Get("/quizzes", quizController.Index)
	api.Middleware(middleware.Auth()).Get("/quizzes/all", quizController.All)
	api.Middleware(middleware.Auth()).Get("/quizzes/my-packages", quizController.MyPackages)
	api.Middleware(middleware.Auth()).Get("/quizzes/:package_id/questions", quizController.GetQuestions)
	api.Middleware(middleware.Auth()).Post("/quizzes/submit-result", quizController.SubmitResults)
	api.Middleware(middleware.Auth()).Get("/quizzes/my-quiz-stats", quizController.MyStats)

	//PAYMENT
	api.Middleware(middleware.Auth()).Post("/payments/initiate-free", paymentController.InitiateFree)

	//RANKINGS
	api.Middleware(middleware.Auth()).Get("/rankings/global", rankingController.GlobalRankings)
	api.Middleware(middleware.Auth()).Get("/rankings/package/:package_id", rankingController.PackageRank)
	api.Middleware(middleware.Auth()).Get("/rankings/my-rank", rankingController.MyRank)

}
