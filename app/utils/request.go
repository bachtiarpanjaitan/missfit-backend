package utils

import "github.com/goravel/framework/contracts/http"

func ValidateRequest(ctx http.Context, rules map[string]string) (map[string]any, any) {
	data := ctx.Request().All()

	// 2. inject field
	for field := range rules {
		if _, ok := data[field]; !ok {
			data[field] = ""
		}
	}

	// 3. validasi
	_, err := ctx.Request().Validate(rules)
	if err != nil {
		return nil, BadRequest(ctx, "validation failed", err.Error())
	}

	return data, nil
}
