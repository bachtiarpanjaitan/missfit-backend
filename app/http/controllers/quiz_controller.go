package controllers

import (
	"missfit/app/dtos"
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/services"
	"missfit/app/utils"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/http"
	"github.com/xuri/excelize/v2"
)

type QuizController struct {
	// Dependent services
	packageService services.PackageServiceInterface
}

func NewQuizController(packageService services.PackageServiceInterface) *QuizController {
	return &QuizController{
		// Inject services
		packageService: packageService,
	}
}

func (r *QuizController) Index(ctx http.Context) http.Response {
	user := utils.User(ctx)

	// Ambil ID paket yang sudah dimiliki user (is_active = true)
	purchasedIds := r.getPurchasedPackageIds(user.Id)

	var paid_packages []models.QuizPackage
	var free_packages []models.QuizPackage

	allowed := models.QuizPackage{}.AllowedFields()

	query := facades.Orm().
		Query().
		Where("is_published = ?", true).
		Where("is_free = ?", false).
		Order("created_at DESC")
	q := utils.ApplyQueryParams(ctx, query, allowed)
	q.Find(&paid_packages)

	query = facades.Orm().
		Query().
		Where("is_published = ?", true).
		Where("is_free = ?", true).
		Order("created_at DESC")
	q = utils.ApplyQueryParams(ctx, query, allowed)
	q.Find(&free_packages)

	// Tandai setiap paket dengan status IsPurchased
	paidWithStatus := dtos.BuildPackageResponses(paid_packages, purchasedIds)
	freeWithStatus := dtos.BuildPackageResponses(free_packages, purchasedIds)
	latestWithStatus := append(paidWithStatus, freeWithStatus...)

	return ctx.Response().Json(200, map[string]any{
		"message": "data loaded",
		"data": map[string]any{
			"paid_packages":   paidWithStatus,
			"free_packages":   freeWithStatus,
			"latest_packages": latestWithStatus,
		},
	})
}

func (r *QuizController) All(ctx http.Context) http.Response {
	user := utils.User(ctx)

	// Ambil ID paket yang sudah dimiliki user (is_active = true)
	purchasedIds := r.getPurchasedPackageIds(user.Id)

	var packages []models.QuizPackage
	query := facades.Orm().
		Query().
		Where("is_published = ?", true).
		Order("created_at DESC")

	allowed := models.QuizPackage{}.AllowedFields()
	q := utils.ApplyQueryParams(ctx, query, allowed)
	q.Find(&packages)

	// Tandai setiap paket dengan status IsPurchased
	result := dtos.BuildPackageResponses(packages, purchasedIds)

	return ctx.Response().Json(200, map[string]any{
		"message": "data loaded",
		"data": map[string]any{
			"packages": result,
		},
	})
}

// getPurchasedPackageIds mengembalikan map berisi quiz_package_id yang sudah
// dibeli/diklaim user dan statusnya aktif (is_active = true).
// Menggunakan map[string]bool untuk lookup O(1) saat menandai IsPurchased.
func (r *QuizController) getPurchasedPackageIds(userId string) map[string]bool {
	var purchased []models.UserPurchasedPackage
	facades.Orm().Query().
		Where("user_id", userId).
		Where("is_active", true).
		Find(&purchased)

	ids := make(map[string]bool, len(purchased))
	for _, p := range purchased {
		ids[p.QuizPackageId] = true
	}
	return ids
}

func (r *QuizController) MyPackages(ctx http.Context) http.Response {
	user := utils.User(ctx)
	pagination := utils.ParsePagination(ctx)

	userPackages, err := r.packageService.GetUserPackages(user.Base.Id, pagination)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}

	return utils.Ok(ctx, "loaded", userPackages)

}

func (r *QuizController) GetQuestions(ctx http.Context) http.Response {
	package_id := ctx.Request().Route("package_id")
	if package_id == "" {
		return utils.BadRequest(ctx, "Paket tidak ditemukan", nil)
	}

	pack, err := r.packageService.GetPackageById(package_id, nil)
	if pack == nil || err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}
	if !pack.IsPublished {
		return utils.BadRequest(ctx, "Paket tidak dipublikasikan", nil)
	}

	questions, err := r.packageService.GetQuestionsByPackageId(package_id)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}

	return utils.Ok(ctx, "loaded", map[string]any{
		"questions": questions,
		"package":   pack,
	})
}

func (r *QuizController) SubmitResults(ctx http.Context) http.Response {
	data, err := utils.ValidateRequest(ctx, map[string]string{
		"packageId":      "required|min_len:1",
		"answers":        "required|array",
		"score":          "required|float",
		"totalQuestions": "required|integer",
		"timeSpent":      "required|integer",
		"startedAt":      "required",
		"completedAt":    "required",
	})
	if err != nil {
		return err.(http.Response)
	}
	user := utils.User(ctx)
	var quizResult dtos.QuizResult

	//reach max attempt
	hasMaxAttempts, err := r.packageService.HasMaxAttempts(user.Id, data["packageId"].(string))
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}
	if hasMaxAttempts {
		return utils.BadRequest(ctx, "Anda telah mencapai batas maksimum kesempatan mengikuti kuis ini", nil)
	}

	answersRaw := data["answers"].([]any)

	var answers []dtos.QuizResultAnswer

	for _, a := range answersRaw {
		item := a.(map[string]any)

		answers = append(answers, dtos.QuizResultAnswer{
			QuestionId: item["questionId"].(string),
			AnswerId:   item["answerId"].(string),
		})
	}

	quizResult.Answers = answers
	quizResult.UserId = user.Id
	quizResult.PackageId = data["packageId"].(string)
	quizResult.Score = data["score"].(float64)
	quizResult.TotalQuestions = data["totalQuestions"].(float64)
	quizResult.TimeSpent = data["timeSpent"].(float64)

	startedAt, err := utils.ToDateTime(data["startedAt"].(string))
	if err != nil {
		return utils.BadRequest(ctx, "Invalid startedAt", err)
	}
	quizResult.StartedAt = startedAt
	completedAt, err := utils.ToDateTime(data["completedAt"].(string))
	if err != nil {
		return utils.BadRequest(ctx, "Invalid completedAt", err)
	}
	quizResult.CompletedAt = completedAt

	result, err := r.packageService.SubmitQuizResult(quizResult, user)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}

	return utils.Ok(ctx, "Success", result)
}

func (r *QuizController) MyStats(ctx http.Context) http.Response {
	user := utils.User(ctx)
	userResults, err := r.packageService.GetUserResults(user.Id)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}

	return utils.Ok(ctx, "loaded", userResults)
}

func (r *QuizController) ImportQuizzes(ctx http.Context) http.Response {
	file, err := ctx.Request().File("file")
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": "File wajib diupload",
			"error":   err.Error(),
		})
	}

	xlsx, err := excelize.OpenFile(file.File())
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": "File excel tidak valid",
			"error":   err.Error(),
		})
	}
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": "File excel tidak valid",
			"error":   err.Error(),
		})
	}

	rows, err := xlsx.GetRows("Sheet1")
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": "Sheet1 tidak ditemukan",
			"error":   err.Error(),
		})
	}

	if len(rows) <= 17 {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": "Data soal kosong",
		})
	}

	packageId := uuid.NewString()

	title, _ := xlsx.GetCellValue("Sheet1", "B1")
	description, _ := xlsx.GetCellValue("Sheet1", "B2")
	category, _ := xlsx.GetCellValue("Sheet1", "B3")
	educationLevel, _ := xlsx.GetCellValue("Sheet1", "B4")
	difficulty, _ := xlsx.GetCellValue("Sheet1", "B5")
	thumbnail, _ := xlsx.GetCellValue("Sheet1", "B6")
	isFreeStr, _ := xlsx.GetCellValue("Sheet1", "B7")
	priceStr, _ := xlsx.GetCellValue("Sheet1", "B8")
	currency, _ := xlsx.GetCellValue("Sheet1", "B9")
	totalQuestionStr, _ := xlsx.GetCellValue("Sheet1", "B10")
	durationStr, _ := xlsx.GetCellValue("Sheet1", "B11")
	passingScoreStr, _ := xlsx.GetCellValue("Sheet1", "B12")
	maxAttemptsStr, _ := xlsx.GetCellValue("Sheet1", "B13")
	isPublishedStr, _ := xlsx.GetCellValue("Sheet1", "B14")

	price, _ := strconv.ParseFloat(priceStr, 64)

	isFree := strings.ToLower(isFreeStr) == "true" ||
		isFreeStr == "1"

	durationMinutes, _ := strconv.Atoi(durationStr)
	passingScore, _ := strconv.Atoi(passingScoreStr)
	maxAttempts, _ := strconv.Atoi(maxAttemptsStr)
	totalQuestion, _ := strconv.Atoi(totalQuestionStr)
	isPublished, _ := strconv.ParseBool(isPublishedStr)

	quizPackage := models.QuizPackage{
		Base: models.Base{
			Id:        packageId,
			CreatedAt: time.Now(),
		},

		Title:           title,
		Description:     description,
		Category:        category,
		DifficultyLevel: difficulty,
		EducationLevel:  educationLevel,
		ThumbnailUrl:    thumbnail,
		Price:           price,
		IsFree:          isFree,
		Currency:        currency,
		DurationMinutes: durationMinutes,
		PassingScore:    passingScore,
		MaxAttempts:     maxAttempts,
		TotalQuestions:  totalQuestion,
		TotalTaken:      0,
		AverageScore:    0,
		Rating:          0.00,
		IsPublished:     isPublished,
	}

	tx, err := facades.DB().BeginTransaction()

	_, err = tx.Table("quiz_packages").Insert(&quizPackage)
	if err != nil {
		tx.Rollback()
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"message": "Gagal membuat quiz package",
			"error":   err.Error(),
		})
	}

	importedQuestions := 0

	// mulai dari row 18
	for i := 17; i < len(rows); i++ {
		row := rows[i]

		// skip row kosong
		if len(row) < 10 {
			continue
		}

		question := ""
		if len(row) > 1 {
			question = strings.TrimSpace(row[1])
		}

		if question == "" {
			continue
		}

		orderNo := 0
		if len(row) > 3 {
			orderNo, _ = strconv.Atoi(strings.TrimSpace(row[2]))
		}

		questionType := "multiple_choice"
		if len(row) > 4 && row[4] != "" {
			questionType = row[4]
		}

		point := 10.0
		if len(row) > 7 {
			point, _ = strconv.ParseFloat(strings.TrimSpace(row[7]), 64)
		}

		quizQuestionId := uuid.NewString()
		QuestionImageUrl := strings.TrimSpace(row[2])
		explanation := strings.TrimSpace(row[6])

		quizQuestion := models.QuizQuestion{
			Base: models.Base{
				Id:        quizQuestionId,
				CreatedAt: time.Now(),
			},

			QuizPackageId:    packageId,
			QuestionText:     question,
			QuestionType:     questionType,
			QuestionImageUrl: QuestionImageUrl,
			Point:            point,
			QuestionOrder:    orderNo,
			Explanation:      explanation,
		}
		_, err = tx.Table("quiz_questions").Insert(&quizQuestion)
		if err != nil {
			tx.Rollback()
			return ctx.Response().Json(http.StatusInternalServerError, http.Json{
				"message": "Gagal insert question",
				"error":   err.Error(),
				"row":     i + 1,
			})
		}

		// option mulai kolom J(index 9)
		// optionOrder := 1

		for col := 9; col < len(row); col += 4 {
			if col >= len(row) {
				break
			}

			optionText := strings.TrimSpace(row[col])

			if optionText == "" {
				continue
			}

			isCorrect := false

			// kolom is_correct = col+3
			if col+3 < len(row) {
				val := strings.TrimSpace(row[col+3])
				if val == "YA" || val == "ya" {
					isCorrect = true
				}
			}

			imgUrl := ""

			// kolom image url = col+1
			if col+3 < len(row) {
				imgUrl = strings.ToLower(strings.TrimSpace(row[col+1]))
			}

			optionOrder, _ := strconv.Atoi(strings.TrimSpace(row[col+2]))

			quizOption := models.QuizOption{
				Base: models.Base{
					Id:        uuid.NewString(),
					CreatedAt: time.Now(),
				},

				QuizQuestionId: quizQuestionId,
				OptionText:     optionText,
				OptionImageUrl: imgUrl,
				OptionOrder:    optionOrder,
				IsCorrect:      isCorrect,
			}

			_, err = tx.Table("quiz_options").Insert(&quizOption)
			if err != nil {
				tx.Rollback()
				return ctx.Response().Json(http.StatusInternalServerError, http.Json{
					"message": "Gagal insert option",
					"error":   err.Error(),
					"row":     i + 1,
				})
			}

			// optionOrder++
		}

		importedQuestions++
	}

	tx.Commit()
	return ctx.Response().Json(http.StatusOK, http.Json{
		"message":            "Berhasil import quiz",
		"quiz_package_id":    packageId,
		"imported_questions": importedQuestions,
	})
}
