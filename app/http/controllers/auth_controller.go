package controllers

import (
	"encoding/json"
	"fmt"
	"lumos/app/facades"
	"lumos/app/models"
	"lumos/app/utils"
	nethttp "net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"lumos/app/services"

	"github.com/goravel/framework/contracts/http"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	packageService services.PackageServiceInterface
}

type googleTokenInfo struct {
	Audience      string `json:"aud"`
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
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

	// Generate verification token (64 character random string)
	verificationToken := utils.GenerateId() + utils.GenerateId()
	// Token expires in 24 hours
	tokenExpiresAt := time.Now().Add(24 * time.Hour)

	appUrl := facades.Config().GetString("app.url")

	user := models.User{
		Name:                            name,
		Email:                           email,
		Username:                        username,
		Password:                        string(hashed),
		Gender:                          gender,
		AuthProvider:                    "local",
		IsActive:                        true,
		IsVerified:                      false,
		Role:                            "user",
		Phone:                           phone,
		AvatarURL:                       appUrl + "/public/uploads/avatar/default.svg",
		EmailVerificationToken:          verificationToken,
		EmailVerificationTokenExpiresAt: &tokenExpiresAt,
	}

	facades.Orm().Query().Create(&user)

	// Send verification email
	go r.sendVerificationEmail(user)

	return ctx.Response().Json(201, map[string]interface{}{
		"message": "Berhasil mendaftar, silahkan verifikasi email anda",
		"data": map[string]interface{}{
			"user": user,
		},
	})
}

// sendVerificationEmail mengirim email verifikasi ke user
func (r *AuthController) sendVerificationEmail(user models.User) {
	appUrl := facades.Config().GetString("app.url")
	verificationLink := fmt.Sprintf("%s/email-verify?token=%s", appUrl, user.EmailVerificationToken)

	// Kirim email menggunakan Goravel Mail
	mail := facades.Mail()

	// Set from address
	fromAddress := facades.Config().GetString("mail.from.address")
	fromName := facades.Config().GetString("mail.from.name")

	// Kirim email sederhana (HTML)
	subject := "Verifikasi Email - MissFit"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Verifikasi Email</title>
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2c3e50;">Verifikasi Email Anda</h2>
				<p>Halo %s,</p>
				<p>Terima kasih telah mendaftar di MissFit. Silakan klik link dibawah untuk verifikasi email Anda:</p>
				<p><a href="%s" style="background: #3498db; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Verifikasi Email</a></p>
				<p>Atau salin link berikut di browser Anda:</p>
				<p style="word-break: break-all;">%s</p>
				<p>Link ini akan kadaluwarsa dalam 24 jam.</p>
				<p>Jika Anda tidak mendaftar, silakan abaikan email ini.</p>
				<p>Salam,<br>Tim MissFit</p>
			</div>
		</body>
		</html>
	`, user.Name, verificationLink, verificationLink)

	// Send email
	err := mail.To([]string{user.Email}).
		From(fromName, fromAddress).
		Subject(subject).
		Html(body).
		Send()

	// Log error jika gagal, tapi tidak mengganggu register
	if err != nil {
		fmt.Printf("Failed to send verification email to %s: %v\n", user.Email, err)
	}
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

func (r *AuthController) GoogleLogin(ctx http.Context) http.Response {
	idToken := strings.TrimSpace(ctx.Request().Input("id_token"))
	if idToken == "" {
		idToken = strings.TrimSpace(ctx.Request().Input("token"))
	}
	if idToken == "" {
		return utils.BadRequest(ctx, "Token Google wajib diisi", nil)
	}

	googleUser, err := verifyGoogleIDToken(idToken)
	if err != nil {
		return utils.BadRequest(ctx, "Token Google tidak valid", err.Error())
	}

	var user models.User
	facades.Orm().Query().
		Where("auth_provider", "google").
		Where("auth_provider_id", googleUser.Subject).
		First(&user)

	if user.Id == "" {
		facades.Orm().Query().
			Where("email", googleUser.Email).
			First(&user)
	}

	now := time.Now()
	if user.Id == "" {
		user = models.User{
			Name:           googleUser.Name,
			Email:          googleUser.Email,
			Username:       generateUniqueGoogleUsername(googleUser.Email, googleUser.Name),
			Password:       hashedGooglePlaceholderPassword(),
			AvatarURL:      googleAvatarURL(googleUser.Picture),
			AuthProvider:   "google",
			AuthProviderID: googleUser.Subject,
			IsActive:       true,
			IsVerified:     true,
			Role:           "user",
			LastLoginAt:    &now,
		}

		if err := facades.Orm().Query().Create(&user); err != nil {
			return utils.InternalServerError(ctx, "Gagal membuat pengguna Google", err.Error())
		}
	} else {
		if !user.IsActive {
			return utils.BadRequest(ctx, "Akunmu sedang tidak aktif", nil)
		}

		updateData := map[string]any{
			"auth_provider":    "google",
			"auth_provider_id": googleUser.Subject,
			"is_verified":      true,
			"last_login_at":    now,
		}

		if user.Name == "" && googleUser.Name != "" {
			updateData["name"] = googleUser.Name
			user.Name = googleUser.Name
		}

		if shouldUseGoogleAvatar(user.AvatarURL) && googleUser.Picture != "" {
			updateData["avatar_url"] = googleUser.Picture
			user.AvatarURL = googleUser.Picture
		}

		_, err := facades.Orm().Query().
			Model(&models.User{}).
			Where("id", user.Id).
			Update(updateData)
		if err != nil {
			return utils.InternalServerError(ctx, "Gagal memperbarui pengguna Google", err.Error())
		}

		user.AuthProvider = "google"
		user.AuthProviderID = googleUser.Subject
		user.IsVerified = true
		user.LastLoginAt = &now
	}

	token, err := utils.GenerateToken(user.Id)
	if err != nil {
		return utils.InternalServerError(ctx, "Gagal membuat token", err.Error())
	}

	return utils.Ok(ctx, "login google success", map[string]any{
		"token": token,
		"user":  user,
	})
}

func verifyGoogleIDToken(idToken string) (*googleTokenInfo, error) {
	clientIDs := configuredGoogleClientIDs()
	if len(clientIDs) == 0 {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID atau GOOGLE_CLIENT_IDS belum dikonfigurasi")
	}

	client := nethttp.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenInfo googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}

	if resp.StatusCode != nethttp.StatusOK {
		return nil, fmt.Errorf("Google tokeninfo status %d", resp.StatusCode)
	}
	if tokenInfo.Subject == "" || tokenInfo.Email == "" {
		return nil, fmt.Errorf("profil Google tidak lengkap")
	}
	if !isGoogleEmailVerified(tokenInfo.EmailVerified) {
		return nil, fmt.Errorf("email Google belum terverifikasi")
	}
	if !isAllowedGoogleAudience(tokenInfo.Audience, clientIDs) {
		return nil, fmt.Errorf("audience Google tidak sesuai")
	}

	return &tokenInfo, nil
}

func configuredGoogleClientIDs() []string {
	keys := []string{
		"GOOGLE_CLIENT_IDS",
		"GOOGLE_CLIENT_ID",
		"GOOGLE_WEB_CLIENT_ID",
		"GOOGLE_ANDROID_CLIENT_ID",
		"GOOGLE_IOS_CLIENT_ID",
	}

	seen := map[string]bool{}
	var clientIDs []string
	for _, key := range keys {
		raw := strings.TrimSpace(fmt.Sprint(facades.Config().Env(key, "")))
		for _, value := range strings.Split(raw, ",") {
			clientID := strings.TrimSpace(value)
			if clientID == "" || seen[clientID] {
				continue
			}
			seen[clientID] = true
			clientIDs = append(clientIDs, clientID)
		}
	}

	return clientIDs
}

func isAllowedGoogleAudience(audience string, allowedClientIDs []string) bool {
	for _, clientID := range allowedClientIDs {
		if audience == clientID {
			return true
		}
	}
	return false
}

func isGoogleEmailVerified(value any) bool {
	switch verified := value.(type) {
	case bool:
		return verified
	case string:
		return strings.EqualFold(verified, "true")
	default:
		return false
	}
}

func generateUniqueGoogleUsername(email string, name string) string {
	base := googleUsernameBase(email, name)
	candidate := base

	for i := 0; i < 10; i++ {
		var existing models.User
		facades.Orm().Query().Where("username", candidate).First(&existing)
		if existing.Id == "" {
			return candidate
		}
		candidate = fmt.Sprintf("%s%d", base, i+1)
	}

	id := utils.GenerateId()
	if len(id) > 8 {
		id = id[:8]
	}
	return base + "_" + id
}

func googleUsernameBase(email string, name string) string {
	source := strings.TrimSpace(strings.Split(email, "@")[0])
	if source == "" {
		source = name
	}
	source = strings.ToLower(source)

	var builder strings.Builder
	for _, char := range source {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		case char == '_' || char == '.':
			builder.WriteRune(char)
		case char == '-' || char == ' ':
			builder.WriteRune('_')
		}
	}

	username := strings.Trim(builder.String(), "._")
	if len(username) < 3 {
		username = "user"
	}
	return username
}

func hashedGooglePlaceholderPassword() string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(utils.GenerateId()), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashed)
}

func googleAvatarURL(picture string) string {
	if picture != "" {
		return picture
	}

	appUrl := facades.Config().GetString("app.url")
	return appUrl + "/public/uploads/avatar/default.svg"
}

func shouldUseGoogleAvatar(currentAvatarURL string) bool {
	return currentAvatarURL == "" || strings.Contains(currentAvatarURL, "/public/uploads/avatar/default.svg")
}

func (r *AuthController) Me(ctx http.Context) http.Response {
	user, errResp := utils.AuthUser(ctx)
	if errResp != nil {
		return errResp
	}

	type userQuizProgressSummary struct {
		TotalPoints           float64 `gorm:"column:total_points"`
		TotalQuizzesCompleted int64   `gorm:"column:total_quizzes_completed"`
	}

	var summary userQuizProgressSummary
	err := facades.Orm().Query().Raw(`
		SELECT
			COALESCE(SUM(latest.total_points), 0) AS total_points,
			COUNT(latest.quiz_package_id) AS total_quizzes_completed
		FROM (
			SELECT DISTINCT ON (quiz_package_id)
				quiz_package_id,
				COALESCE(total_points, 0) AS total_points
			FROM user_quiz_attempts
			WHERE user_id = ?
				AND deleted_at IS NULL
				AND completed_at IS NOT NULL
			ORDER BY quiz_package_id, created_at DESC
		) latest
	`, user.Id).Scan(&summary)
	if err != nil {
		return utils.InternalServerError(ctx, "Gagal menghitung ulang progress pengguna", err.Error())
	}

	totalQuizzesCompleted := int(summary.TotalQuizzesCompleted)
	if user.TotalPoints != summary.TotalPoints || user.TotalQuizzesCompleted != totalQuizzesCompleted {
		_, err = facades.Orm().Query().
			Model(&models.User{}).
			Where("id", user.Id).
			Update(map[string]any{
				"total_points":            summary.TotalPoints,
				"total_quizzes_completed": totalQuizzesCompleted,
			})
		if err != nil {
			return utils.InternalServerError(ctx, "Gagal memperbarui progress pengguna", err.Error())
		}

		user.TotalPoints = summary.TotalPoints
		user.TotalQuizzesCompleted = totalQuizzesCompleted
	}

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

func (r *AuthController) ViewDeleteAccount(ctx http.Context) http.Response {
	email := ctx.Request().Route("email")
	return ctx.Response().View().Make("delete_account.tmpl", map[string]any{
		"title": "Hapus Akun",
		"email": email,
		"id":    utils.GenerateId(),
	})
}

func (r *AuthController) DeleteAccount(ctx http.Context) http.Response {
	rawData := ctx.Request().All()
	var errMsg string
	if rawData["id"].(string) != rawData["confirm"].(string) {
		errMsg = "Data Konfirmasi Salah, Silahkan masukkan data konfirmasi dengan benar."
	}

	var dbUser models.User
	dbErr := facades.Orm().Query().Where("email", rawData["email"].(string)).First(&dbUser)
	if dbErr != nil || dbUser.Id == "" {
		errMsg = "Pengguna tidak ditemukan"
		return ctx.Response().View().Make("delete_account.tmpl", map[string]any{
			"email": rawData["email"].(string),
			"error": errMsg,
			"id":    utils.GenerateId(),
		})
	}

	if errBcrypt := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(rawData["password"].(string))); errBcrypt != nil {
		errMsg = "Password anda tidak benar, masukkan password lama anda."
		return ctx.Response().View().Make("delete_account.tmpl", map[string]any{
			"email": rawData["email"].(string),
			"error": errMsg,
			"id":    utils.GenerateId(),
		})
	}

	if errMsg != "" {
		return ctx.Response().View().Make("delete_account.tmpl", map[string]any{
			"email": rawData["email"].(string),
			"error": errMsg,
			"id":    utils.GenerateId(),
		})
	}

	_, errUpdate := facades.Orm().Query().Model(&models.User{}).Where("id", dbUser.Id).Update(map[string]any{
		"deleted_at": time.Now(),
		"is_active":  false,
	})
	if errUpdate != nil {
		errMsg = "Gagal menghapus data pengguna, silakan coba lagi."
		return ctx.Response().View().Make("delete_account.tmpl", map[string]any{
			"email": rawData["email"].(string),
			"error": errMsg,
			"id":    utils.GenerateId(),
		})
	}

	return ctx.Response().View().Make("success_delete_account.tmpl")
}

// ViewEmailVerify menampilkan halaman verifikasi email
func (r *AuthController) ViewEmailVerify(ctx http.Context) http.Response {
	token := ctx.Request().Query("token")

	if token == "" {
		return ctx.Response().View().Make("email_verify_error.tmpl", map[string]any{
			"error": "Token verifikasi tidak ditemukan",
		})
	}

	var user models.User
	err := facades.Orm().Query().
		Where("email_verification_token", token).
		First(&user)

	if err != nil || user.Id == "" {
		return ctx.Response().View().Make("email_verify_error.tmpl", map[string]any{
			"error": "Token verifikasi tidak valid atau sudah kadaluwarsa",
		})
	}

	// Check if token is expired
	if user.EmailVerificationTokenExpiresAt != nil && time.Now().After(*user.EmailVerificationTokenExpiresAt) {
		return ctx.Response().View().Make("email_verify_error.tmpl", map[string]any{
			"error": "Token verifikasi sudah kadaluwarsa. Silakan daftar ulang.",
		})
	}

	// Token valid, verify user
	_, err = facades.Orm().Query().
		Model(&models.User{}).
		Where("id", user.Id).
		Update(map[string]any{
			"is_verified":                         true,
			"email_verification_token":            nil,
			"email_verification_token_expires_at": nil,
		})

	if err != nil {
		return ctx.Response().View().Make("email_verify_error.tmpl", map[string]any{
			"error": "Gagal memverifikasi akun. Silakan coba lagi.",
		})
	}

	return ctx.Response().View().Make("email_verify_success.tmpl", map[string]any{
		"message": "Email berhasil diverifikasi! Silakan login untuk melanjutkan.",
	})
}

// VerifyEmail API endpoint untuk verifikasi email (alternative JSON response)
func (r *AuthController) VerifyEmail(ctx http.Context) http.Response {
	token := ctx.Request().Query("token")

	if token == "" {
		return utils.BadRequest(ctx, "Token verifikasi wajib diisi", nil)
	}

	var user models.User
	err := facades.Orm().Query().
		Where("email_verification_token", token).
		First(&user)

	if err != nil || user.Id == "" {
		return utils.BadRequest(ctx, "Token verifikasi tidak valid atau sudah kadaluwarsa", nil)
	}

	// Check if token is expired
	if user.EmailVerificationTokenExpiresAt != nil && time.Now().After(*user.EmailVerificationTokenExpiresAt) {
		return utils.BadRequest(ctx, "Token verifikasi sudah kadaluwarsa. Silakan daftar ulang.", nil)
	}

	// Token valid, verify user
	_, err = facades.Orm().Query().
		Model(&models.User{}).
		Where("id", user.Id).
		Update(map[string]any{
			"is_verified":                         true,
			"email_verification_token":            nil,
			"email_verification_token_expires_at": nil,
		})

	if err != nil {
		return utils.InternalServerError(ctx, "Gagal memverifikasi akun", err)
	}

	return utils.Ok(ctx, "Email berhasil diverifikasi", nil)
}
