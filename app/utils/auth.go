package utils

import (
	"lumos/app/models"

	"github.com/goravel/framework/contracts/http"
)

func AuthUser(ctx http.Context) (*models.User, http.Response) {
	userRaw := ctx.Value("user")
	if userRaw == nil {
		return nil, ctx.Response().Json(401, map[string]any{
			"message": "unauthorized",
		})
	}

	user, ok := userRaw.(*models.User)
	if !ok {
		return nil, ctx.Response().Json(401, map[string]any{
			"message": "unauthorized",
		})
	}

	return user, nil
}

func User(ctx http.Context) *models.User {
	user := ctx.Value("user").(*models.User)
	if user == nil {
		return nil
	}

	return user
}
