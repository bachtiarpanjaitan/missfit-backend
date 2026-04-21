package utils

import (
	"github.com/goravel/framework/contracts/http"
)

func Ok(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Json(http.StatusOK, map[string]any{
		"message": message,
		"data":    data,
	})
}

func BadRequest(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Json(http.StatusBadRequest, map[string]any{
		"message": message,
		"data":    data,
	})
}

func Unauthorized(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Json(http.StatusUnauthorized, map[string]any{
		"message": message,
		"data":    data,
	})
}

func Forbidden(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Json(http.StatusForbidden, map[string]any{
		"message": message,
		"data":    data,
	})
}

func NotFound(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Json(http.StatusNotFound, map[string]any{
		"message": message,
		"data":    data,
	})
}

func InternalServerError(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Json(http.StatusInternalServerError, map[string]any{
		"message": message,
		"data":    data,
	})
}
