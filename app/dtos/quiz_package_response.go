package dtos

import "lumos/app/models"

// QuizPackageResponse membungkus QuizPackage dengan informasi status pembelian user.
// Digunakan di endpoint /quizzes/all dan /quizzes agar frontend tahu
// paket mana yang sudah dimiliki user yang sedang login.
type QuizPackageResponse struct {
	models.QuizPackage

	// IsPurchased true jika user sudah membeli/klaim paket ini dan is_active = true.
	// Frontend menggunakan field ini untuk menyembunyikan paket dari Packages screen
	// dan menampilkan tombol "Mulai Kuis" di tempat lain.
	IsPurchased bool `json:"IsPurchased"`
}

// BuildPackageResponses mengambil daftar paket dan set ID paket yang sudah dibeli,
// lalu menggabungkannya menjadi slice QuizPackageResponse.
func BuildPackageResponses(packages []models.QuizPackage, purchasedIds map[string]bool) []QuizPackageResponse {
	result := make([]QuizPackageResponse, 0, len(packages))
	for _, pkg := range packages {
		result = append(result, QuizPackageResponse{
			QuizPackage: pkg,
			IsPurchased: purchasedIds[pkg.Id],
		})
	}
	return result
}
