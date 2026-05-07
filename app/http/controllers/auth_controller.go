package controllers

import (
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"
	"time"

	"missfit/app/services"

	"github.com/goravel/framework/contracts/http"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	packageService services.PackageServiceInterface
}

func NewAuthController(packageService services.PackageServiceInterface) *AuthController {
	return &AuthController{
		packageService: packageService,
	}
}

func (r *AuthController) Index(ctx http.Context) http.Response {
	return nil
}

func (r *AuthController) Register(ctx http.Context) http.Response {
	email := ctx.Request().Input("email")
	name := ctx.Request().Input("name")
	password := ctx.Request().Input("password")
	confirmPassword := ctx.Request().Input("confirm_password")
	gender := ctx.Request().Input("gender")
	username := ctx.Request().Input("username")
	phone := ctx.Request().Input("phone")

	if confirmPassword != password {
		return ctx.Response().Json(400, "password dan konfirmasi password tidak sama")
	}

	if email == "" || password == "" || username == "" {
		return ctx.Response().Json(400, "email, username, password wajib")
	}

	var existing models.User
	facades.Orm().Query().
		Where("email", email).
		OrWhere("username", username).
		First(&existing)

	if existing.Id != "" {
		return ctx.Response().Json(400, "username atau email sudah pernah digunakan")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Name:         name,
		Email:        email,
		Username:     username,
		Password:     string(hashed),
		Gender:       gender,
		AuthProvider: "local",
		IsActive:     true,
		IsVerified:   false,
		Role:         "user",
		Phone:        phone,
	}

	appUrl := facades.Config().GetString("app.url")

	user.AvatarURL = appUrl + "/public/uploads/avatar/default.svg"

	facades.Orm().Query().Create(&user)

	token, _ := utils.GenerateToken(user.Id)

	return ctx.Response().Json(201, map[string]interface{}{
		"message": "Berhasil mendaftar, silahkan verifikasi email anda",
		"data": map[string]interface{}{
			"token": token,
			"user":  user,
		},
	})
}

func (r *AuthController) Login(ctx http.Context) http.Response {
	email := ctx.Request().Input("email")
	password := ctx.Request().Input("password")

	// validasi basic (biar gak kirim kosong terus berharap keajaiban)
	if email == "" || password == "" {
		return ctx.Response().Json(400, map[string]string{
			"error": "email dan password wajib",
		})
	}

	var user models.User

	if !user.IsActive {
		utils.BadRequest(ctx, "Akunmu sedang tidak aktif", nil)
	}

	if !user.IsVerified {
		utils.BadRequest(ctx, "Akunmu belum terverifikasi, silakan verifikasi terlebih dahulu", nil)
	}

	// ambil user berdasarkan email
	err := facades.Orm().Query().
		Where("email", email).
		First(&user)

	if err != nil || user.Id == "" {
		return ctx.Response().Json(401, map[string]string{
			"error": "invalid credentials",
		})
	}

	// cek password
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return ctx.Response().Json(401, map[string]string{
			"error": "invalid credentials",
		})
	}

	// update last login
	now := time.Now()

	facades.Orm().Query().
		Model(&models.User{}).
		Where("id", user.Id).
		Update("last_login_at", now)

	// generate token
	token, err := utils.GenerateToken(user.Id)
	if err != nil {
		return utils.InternalServerError(ctx, "Gagal membuat token", err)
	}

	return utils.Ok(ctx, "login success", map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":                      user.Id,
				"email":                   user.Email,
				"name":                    user.Name,
				"username":                user.Username,
				"avatar_url":              user.AvatarURL,
				"is_active":               user.IsActive,
				"is_verified":             user.IsVerified,
				"last_login_at":           user.LastLoginAt,
				"auth_provider":           user.AuthProvider,
				"total_points":            user.TotalPoints,
				"total_quizzes_completed": user.TotalQuizzesCompleted,
			},
		},
	})
}

func (r *AuthController) Me(ctx http.Context) http.Response {
	user := utils.User(ctx)

	return utils.Ok(ctx, "success", map[string]interface{}{
		"user":  user,
		"token": ctx.Request().Header("Authorization"),
	})
}
