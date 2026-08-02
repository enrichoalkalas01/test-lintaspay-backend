package filter

import "strings"

type BaseFilter[F any] struct {
	// Pagination
	Page  int `query:"page" validate:"min=1"`
	Limit int `query:"limit" validate:"min=1,max=100"`

	// Sorting
	SortBy    string `query:"sort_by"`    // field name
	SortOrder string `query:"sort_order"` // "asc" atau "desc"

	// Searching
	Search string `query:"search"`

	// Filtering fields
	Filters F `query:"filters"` // Implementasi specific akan embed ini
}

// ============ Sorting & Pagination Helpers ============
func (f *BaseFilter[F]) GetDefaultPage() int {
	if f.Page < 1 {
		return 1
	}
	return f.Page
}

func (f *BaseFilter[F]) GetDefaultLimit() int {
	if f.Limit < 1 {
		return 10
	}
	if f.Limit > 100 {
		return 100
	}
	return f.Limit
}

func (f *BaseFilter[F]) GetSortOrder() string {
	order := strings.ToLower(f.SortOrder)
	if order != "asc" && order != "desc" {
		return "desc"
	}
	return order
}

func (f *BaseFilter[F]) GetSortBy() string {
	return strings.TrimSpace(f.SortBy)
}

func (f *BaseFilter[F]) GetSearch() string {
	return strings.TrimSpace(f.Search)
}

func (f *BaseFilter[F]) GetOffset() int {
	return (f.GetDefaultPage() - 1) * f.GetDefaultLimit()
}

// ============ Validation ============
func (f *BaseFilter[F]) ValidatePagination() error {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 10
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	return nil
}

// ============ Query Builder Helpers ============
type QueryCondition struct {
	Field    string
	Operator string // =, !=, >, <, >=, <=, LIKE, IN
	Value    any
}

func BuildQueryConditions(filter any) []QueryCondition {
	return []QueryCondition{}
}
