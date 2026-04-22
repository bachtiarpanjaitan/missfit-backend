package dtos

type PaginationParams struct {
	Limit int    `form:"_limit"`
	Page  int    `form:"_page"`
	Sort  string `form:"_sort"`
	Order string `form:"_order"`
}
