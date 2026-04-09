package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support"

	"missfit/app/facades"
	"missfit/app/http/controllers"
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

	// controllers
	// userController := controllers.NewUserController()
	authController := &controllers.AuthController{}

	// ========================
	// AUTH ROUTES
	// ========================
	facades.Route().Post("/register", authController.Register)
	facades.Route().Post("/login", authController.Login)
}
