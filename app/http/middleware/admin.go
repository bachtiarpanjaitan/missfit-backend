package middleware

import (
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"
	"strings"

	"github.com/goravel/framework/contracts/http"
)

func Admin() http.Middleware {
	return func(ctx http.Context) {
		authHeader := ctx.Request().Header("Authorization")

		if authHeader == "" {
			ctx.Request().AbortWithStatusJson(401, map[string]any{
				"message": "unauthorized",
			})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.Request().AbortWithStatusJson(401, map[string]any{
				"message": "unauthorized",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := utils.ParseToken(tokenString)
		if err != nil {
			ctx.Request().AbortWithStatusJson(401, map[string]any{
				"message": "unauthorized",
			})
			return
		}

		var user models.User
		err = facades.Orm().Query().Where("id", userID).First(&user)
		if err != nil {
			ctx.Request().AbortWithStatusJson(401, map[string]any{
				"message": "unauthorized",
			})
			return
		}

		if user.Id == "" || user.Id == "0" || user.Id == "null" {
			ctx.Request().AbortWithStatusJson(401, map[string]any{
				"message": "unauthorized",
			})
			return
		}

		if user.Role != "admin" {
			ctx.Request().AbortWithStatusJson(401, map[string]any{
				"message": "unauthorized",
			})
			return
		}

		ctx.WithValue(userKey, &user)
		ctx.Request().Next()
	}
}
