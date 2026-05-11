package controllers

import (
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"
	"os"
	"strings"
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
		return utils.BadRequest(ctx, "Email atau password tidak ditemukan", nil)
	}

	var user models.User

	// ambil user berdasarkan email
	err := facades.Orm().Query().
		Where("email", email).
		First(&user)

	if err != nil || user.Id == "" {
		return utils.BadRequest(ctx, "Pengguna tidak ditemukan", nil)
	}

	if !user.IsActive {
		return utils.BadRequest(ctx, "Akunmu sedang tidak aktif", nil)
	}

	if !user.IsVerified {
		return utils.BadRequest(ctx, "Akunmu belum terverifikasi, silakan verifikasi terlebih dahulu", nil)
	}

	// cek password
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return utils.BadRequest(ctx, "Email atau Password salah", nil)
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
				"gender":                  user.Gender,
				"phone":                   user.Phone,
				"bio":                     user.Bio,
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

func (r *AuthController) UpdateProfile(ctx http.Context) http.Response {
	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	all := ctx.Request().All()

	updateData := map[string]any{}

	if name, ok := all["Name"]; ok && name != "" {
		updateData["name"] = name
	}
	if bio, ok := all["Bio"]; ok && bio != "" {
		updateData["bio"] = bio
	}
	if phone, ok := all["Phone"]; ok && phone != "" {
		updateData["phone"] = phone
	}
	if gender, ok := all["Gender"]; ok && gender != "" {
		updateData["gender"] = gender
	}

	// Coba juga key lowercase (fallback untuk JSON body)
	if len(updateData) == 0 {
		if name, ok := all["name"]; ok && name != "" {
			updateData["name"] = name
		}
		if bio, ok := all["bio"]; ok && bio != "" {
			updateData["bio"] = bio
		}
		if phone, ok := all["phone"]; ok && phone != "" {
			updateData["phone"] = phone
		}
		if gender, ok := all["gender"]; ok && gender != "" {
			updateData["gender"] = gender
		}
	}

	if len(updateData) == 0 {
		return utils.BadRequest(ctx, "Tidak ada data yang diperbarui", nil)
	}

	updateData["updated_at"] = time.Now()

	_, err := facades.Orm().Query().Model(&models.User{}).Where("id", user.Id).Update(updateData)
	if err != nil {
		return utils.InternalServerError(ctx, "Gagal memperbarui profil", err.Error())
	}

	var updatedUser models.User
	if err := facades.Orm().Query().Where("id", user.Id).First(&updatedUser); err != nil {
		return utils.InternalServerError(ctx, "Gagal mengambil data profil terbaru", err.Error())
	}

	return utils.Ok(ctx, "Profil berhasil diperbarui", updatedUser)
}

func (r *AuthController) UploadAvatar(ctx http.Context) http.Response {
	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	fileObj, err := ctx.Request().File("avatar")
	if err != nil {
		return utils.BadRequest(ctx, "File avatar tidak ditemukan", nil)
	}

	ext := strings.ToLower(fileObj.GetClientOriginalExtension())
	if ext == "" {
		ext = "jpg"
	}

	allowedExts := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"webp": true,
	}
	if !allowedExts[ext] {
		return utils.BadRequest(ctx, "Format file tidak didukung. Gunakan jpg, jpeg, png, atau webp", nil)
	}

	// Size() mengembalikan (int64, error) di Goravel v1.17
	fileSize, sizeErr := fileObj.Size()
	if sizeErr != nil {
		return utils.InternalServerError(ctx, "Gagal memeriksa ukuran file", sizeErr.Error())
	}
	if fileSize > 5*1024*1024 {
		return utils.BadRequest(ctx, "Ukuran file maksimum 5MB", nil)
	}

	if err := os.MkdirAll("./public/uploads/avatar", 0755); err != nil {
		return utils.InternalServerError(ctx, "Gagal menyiapkan direktori upload", err.Error())
	}

	filename := utils.GenerateId() + "." + ext

	// File() di Goravel v1.17 mengembalikan path string ke file sementara
	tmpPath := fileObj.File()

	content, readErr := os.ReadFile(tmpPath)
	if readErr != nil {
		return utils.InternalServerError(ctx, "Gagal membaca konten file", readErr.Error())
	}

	if err := os.WriteFile("./public/uploads/avatar/"+filename, content, 0644); err != nil {
		return utils.InternalServerError(ctx, "Gagal menyimpan file avatar", err.Error())
	}

	appUrl := facades.Config().GetString("app.url")
	avatarURL := appUrl + "/public/uploads/avatar/" + filename

	_, dbErr := facades.Orm().Query().Model(&models.User{}).Where("id", user.Id).Update("avatar_url", avatarURL)
	if dbErr != nil {
		// Hapus file yang sudah tersimpan jika update DB gagal
		_ = os.Remove("./public/uploads/avatar/" + filename)
		return utils.InternalServerError(ctx, "Gagal memperbarui foto profil di database", dbErr.Error())
	}

	user.AvatarURL = avatarURL

	return utils.Ok(ctx, "Foto profil berhasil diperbarui", user)
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

// ChangePassword memungkinkan user mengganti password mereka.
// Memerlukan: current_password, new_password, confirm_password.
func (r *AuthController) ChangePassword(ctx http.Context) http.Response {
	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	data, err := utils.ValidateRequest(ctx, map[string]string{
		"current_password": "required|min_len:8",
		"new_password":     "required|min_len:8",
		"confirm_password": "required|min_len:8",
	})
	if err != nil {
		return err.(http.Response)
	}

	currentPassword := data["current_password"].(string)
	newPassword := data["new_password"].(string)
	confirmPassword := data["confirm_password"].(string)

	// Validasi: password baru dan konfirmasi harus sama
	if newPassword != confirmPassword {
		return utils.BadRequest(ctx, "Password baru dan konfirmasi password tidak sama", nil)
	}

	// Validasi: password baru tidak boleh sama dengan password lama
	if currentPassword == newPassword {
		return utils.BadRequest(ctx, "Password baru tidak boleh sama dengan password lama", nil)
	}

	// Ambil user dari DB untuk mendapatkan hash password (karena model dari context mungkin tidak include password)
	var dbUser models.User
	dbErr := facades.Orm().Query().Where("id", user.Id).First(&dbUser)
	if dbErr != nil || dbUser.Id == "" {
		return utils.InternalServerError(ctx, "Gagal mengambil data pengguna", nil)
	}

	// Verifikasi password lama
	if errBcrypt := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(currentPassword)); errBcrypt != nil {
		return utils.BadRequest(ctx, "Password lama salah", nil)
	}

	// Hash password baru
	hashed, errHash := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if errHash != nil {
		return utils.InternalServerError(ctx, "Gagal mengenkripsi password", errHash.Error())
	}

	// Update password di database
	_, errUpdate := facades.Orm().Query().Model(&models.User{}).Where("id", user.Id).Update("password", string(hashed))
	if errUpdate != nil {
		return utils.InternalServerError(ctx, "Gagal memperbarui password", errUpdate.Error())
	}

	return utils.Ok(ctx, "Password berhasil diperbarui", nil)
}
