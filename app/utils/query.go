package utils

import (
	"regexp"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/database/orm"
)

// operator mapping
var Operators = map[string]string{
	"eq":   "=",
	"ne":   "!=",
	"gt":   ">",
	"gte":  ">=",
	"lt":   "<",
	"lte":  "<=",
	"like": "LIKE",
	"in":   "IN",
}

// param yang tidak ikut filter
var ReservedParams = map[string]bool{
	"_limit": true,
	"_page":  true,
}

// parse field[operator]
func parseKey(key string) (string, string) {
	re := regexp.MustCompile(`(.*)\[(.*)\]`)
	matches := re.FindStringSubmatch(key)

	if len(matches) == 3 {
		return matches[1], matches[2]
	}

	return key, "eq"
}

func ApplyQueryParams(ctx http.Context, query orm.Query, allowedFields map[string]bool) orm.Query {
	queries := ctx.Request().Queries()

	// ========================
	// FILTER
	// ========================
	for key, values := range queries {
		value := values[0]
		strValue := string(value)
		if ReservedParams[key] {
			continue
		}

		field, op := parseKey(key)

		// whitelist field
		if len(allowedFields) > 0 {
			if !allowedFields[field] {
				continue
			}
		}

		sqlOp, ok := Operators[op]
		if !ok {
			continue
		}

		switch op {
			case "like":
				query = query.Where(field+" "+sqlOp+" ?", "%"+strValue+"%")
			case "in":
				query = query.Where(field+" IN (?)", strings.Split(strValue, ","))
			default:
				query = query.Where(field+" "+sqlOp+" ?", value)
		}
	}

	// ========================
	// PAGINATION
	// ========================
	limit := ctx.Request().QueryInt("_limit", 10)
	page := ctx.Request().QueryInt("_page", 1)

	// validasi biar tidak liar
	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	query = query.Limit(limit).Offset(offset)

	return query
}