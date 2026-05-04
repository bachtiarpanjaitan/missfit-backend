package controllers

import (
	"missfit/app/dtos"
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/services"
	"missfit/app/utils"

	"github.com/goravel/framework/contracts/http"
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
	var paid_packages []models.QuizPackage
	var free_packages []models.QuizPackage
	query := facades.Orm().
		Query().
		With("Questions").
		Where("is_published = ?", true).
		Where("is_free = ?", false).
		Order("created_at DESC")

	allowed := models.QuizPackage{}.AllowedFields()
	q := utils.ApplyQueryParams(ctx, query, allowed)
	q.Find(&paid_packages)

	query = facades.Orm().
		Query().
		With("Questions").
		Where("is_published = ?", true).
		Where("is_free = ?", true).
		Order("created_at DESC")

	q = utils.ApplyQueryParams(ctx, query, allowed)
	q.Find(&free_packages)

	return ctx.Response().Json(200, map[string]any{
		"message": "data loaded",
		"data": map[string]any{
			"paid_packages":   paid_packages,
			"free_packages":   free_packages,
			"latest_packages": append(paid_packages, free_packages...),
		},
	})
}

func (r *QuizController) All(ctx http.Context) http.Response {
	var packages []models.QuizPackage
	query := facades.Orm().
		Query().
		With("Questions").
		Where("is_published = ?", true).
		Order("created_at DESC")

	allowed := models.QuizPackage{}.AllowedFields()
	q := utils.ApplyQueryParams(ctx, query, allowed)
	q.Find(&packages)

	return ctx.Response().Json(200, map[string]any{
		"message": "data loaded",
		"data": map[string]any{
			"packages": packages,
		},
	})
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
