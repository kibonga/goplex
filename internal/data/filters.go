package data

import "goplex.kibonga/internal/validator"

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
