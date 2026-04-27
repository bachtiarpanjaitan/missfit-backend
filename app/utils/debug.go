package utils

import (
	"encoding/json"
	"fmt"

	"github.com/goravel/framework/contracts/http"
)

func Dd(data any) {
	fmt.Println(data)
	panic("stop")
}

func DdResponseJson(ctx http.Context, data any) http.Response {
	return ctx.Response().Json(200, map[string]any{
		"message": "data debug",
		"data":    data,
	})
}

func ToJson(data any) string {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}
