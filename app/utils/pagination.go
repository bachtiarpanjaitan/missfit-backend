package utils

import (
	"lumos/app/dtos"

	"github.com/goravel/framework/contracts/http"
)

func ParsePagination(ctx http.Context) dtos.PaginationParams {
	p := dtos.PaginationParams{
		Limit: ctx.Request().QueryInt("_limit", 10),
		Page:  ctx.Request().QueryInt("_page", 1),
		Sort:  ctx.Request().Query("_sort", "created_at"),
		Order: ctx.Request().Query("_order", "desc"),
	}

	// validasi
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.Page <= 0 {
		p.Page = 1
	}

	return p
}
