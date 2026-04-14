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
