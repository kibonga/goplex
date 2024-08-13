package data

import (
	"math"
	"strings"

	"goplex.kibonga/internal/validator"
)

const (
	MIN_PAGE      int = 0
	MAX_PAGE      int = 10_000_000
	MIN_PAGE_SIZE int = 0
	MAX_PAGE_SIZE int = 100
)

type Filters struct {
	Page            int
	PageSize        int
	Sort            string
	ValidSortValues []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

func ValidateFilters(v *validator.Validator, f *Filters) {
	validatePage(v, f.Page)
	validatePageSize(v, f.PageSize)
	validateSort(v, f.Sort, f.ValidSortValues)
}

func validatePage(v *validator.Validator, page int) {
	v.Check(validator.GreaterThan(page, MIN_PAGE), "page", "must be greater than zero")
	v.Check(!validator.GreaterThan(page, MAX_PAGE), "page", "must be less than 10 million")
}

func validatePageSize(v *validator.Validator, pageSize int) {
	v.Check(validator.GreaterThan(pageSize, MIN_PAGE_SIZE), "page_size", "must be greater than zero")
	v.Check(!validator.GreaterThan(pageSize, MAX_PAGE_SIZE), "page_size", "must be less than one hundred")
}

func validateSort(v *validator.Validator, sortVal string, validSorts []string) {
	v.Check(v.In(sortVal, validSorts...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	for _, v := range f.ValidSortValues {
		if f.Sort == v {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("invalid sort param: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}
