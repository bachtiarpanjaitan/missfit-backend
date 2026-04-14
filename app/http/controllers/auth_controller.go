package controllers

import (
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"
	"time"

	"github.com/goravel/framework/contracts/http"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	// Dependent services
}

func NewAuthController() *AuthController {
	return &AuthController{
		// Inject services
	}
}

func (r *AuthController) Index(ctx http.Context) http.Response {
	return nil
}

func (r *AuthController) Register(ctx http.Context) http.Response {
	email := ctx.Request().Input("email")
	password := ctx.Request().Input("password")
	username := ctx.Request().Input("username")

	if email == "" || password == "" || username == "" {
		return ctx.Response().Json(400, "email, username, password wajib")
	}

	var existing models.User
	facades.Orm().Query().
		Where("email", email).
		OrWhere("username", username).
		First(&existing)

	if existing.Id != "" {
		return ctx.Response().Json(400, "user sudah ada")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// optional fields
	name := ctx.Request().Input("name")
	gender := ctx.Request().Input("gender")

	user := models.User{
		Email:        email,
		Username:     username,
		Password:     string(hashed),
		Name:         name,
		Gender:       gender,
		AuthProvider: "local",
		IsActive:     true,
		IsVerified:   false,
	}

	// avatar default biar gak kosong kayak harapan hidup dev
	user.AvatarURL = "https://ui-avatars.com/api/?name=" + username

	facades.Orm().Query().Create(&user)

	token, _ := utils.GenerateToken(user.Id)

	return ctx.Response().Json(201, map[string]interface{}{
		"message": "register success",
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

	// cek user aktif
	if !user.IsActive {
		return ctx.Response().Json(403, map[string]string{
			"error": "user tidak aktif",
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
		return ctx.Response().Json(500, map[string]string{
			"error": "gagal generate token",
		})
	}

	return ctx.Response().Json(200, map[string]interface{}{
		"message": "login success",
		"data": map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":                      user.Id,
				"email":                   user.Email,
				"name":                    user.Name,
				"username":                user.Username,
				"avatar_url":              user.AvatarURL,
				"gender":                  user.Gender,
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
	userRaw := ctx.Value("user")
	if userRaw == nil {
		return ctx.Response().Json(401, map[string]interface{}{
			"message": "unauthorized",
		})
	}

	user := userRaw.(*models.User)

	return ctx.Response().Json(200, map[string]interface{}{
		"message": "success",
		"data": map[string]interface{}{
			"id":         user.Id,
			"email":      user.Email,
			"name":       user.Name,
			"username":   user.Username,
			"avatar_url": user.AvatarURL,
		},
	})
}
