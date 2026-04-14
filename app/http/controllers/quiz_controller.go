package controllers

import (
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"

	"github.com/goravel/framework/contracts/http"
)

type QuizController struct {
	// Dependent services
}

func NewQuizController() *QuizController {
	return &QuizController{
		// Inject services
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
