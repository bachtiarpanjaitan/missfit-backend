package middleware

import (
	"strings"

	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"

	"github.com/goravel/framework/contracts/http"
)

const userKey = "user"

func Auth() http.Middleware {
	return func(ctx http.Context) {
		authHeader := ctx.Request().Header("Authorization")

		if authHeader == "" {
			ctx.Response().Json(401, map[string]any{
				"message": "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.Response().Json(401, map[string]any{
				"message": "invalid token format",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := utils.ParseToken(tokenString)
		if err != nil {
			ctx.Response().Json(401, map[string]any{
				"message": "invalid or expired token",
			})
			return
		}

		var user models.User
		err = facades.Orm().Query().Where("id", userID).First(&user)
		if err != nil {
			ctx.Response().Json(401, map[string]any{
				"message": "user not found",
			})
			return
		}

		ctx.WithValue(userKey, &user)

		ctx.Request().Next()
	}
}
