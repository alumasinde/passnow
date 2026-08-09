package httpx

import (
	"net/http"
	"strconv"
)

const (
	DefaultPageSize = 25
	MaxPageSize     = 100
)

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ParsePagination reads ?limit=&offset= with safe defaults/caps. Every list
// endpoint should use this instead of parsing query params by hand.
func ParsePagination(r *http.Request) Pagination {
	limit := DefaultPageSize
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	return Pagination{Limit: limit, Offset: offset}
}

type ListEnvelope[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}
