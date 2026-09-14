package query

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type Params struct {
	Page       int
	Limit      int
	Search     string
	Role       string
	IsVerified *bool
	IsActive   *bool
	SortBy     string
	SortOrder  string
}

type Meta struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"totalPages"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
}

func Parse(values url.Values) Params {
	p := Params{
		Page:      parseInt(values.Get("page"), DefaultPage),
		Limit:     parseInt(values.Get("limit"), DefaultLimit),
		Search:    strings.TrimSpace(values.Get("search")),
		Role:      strings.TrimSpace(values.Get("role")),
		SortBy:    strings.TrimSpace(values.Get("sortBy")),
		SortOrder: strings.ToLower(strings.TrimSpace(values.Get("sortOrder"))),
	}
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	if p.SortBy == "" {
		p.SortBy = "createdAt"
	}
	if p.SortOrder != "asc" {
		p.SortOrder = "desc"
	}
	p.IsVerified = parseOptionalBool(values.Get("isVerified"))
	p.IsActive = parseOptionalBool(values.Get("isActive"))
	return p
}

func (p Params) Offset() int { return (p.Page - 1) * p.Limit }

func NewMeta(page, limit int, total int64) Meta {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return Meta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}

func parseInt(v string, fallback int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseOptionalBool(v string) *bool {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}
