package application

import (
	"net/http"
	"strconv"
	"strings"
)

// Params provides convenient access to request parameters
type Params struct {
	r *http.Request
}

// NewParams creates a parameter helper for the request
func NewParams(r *http.Request) *Params {
	return &Params{r: r}
}

// String gets a string parameter from query or form
func (p *Params) String(name string, defaultValue string) string {
	if value := p.r.URL.Query().Get(name); value != "" {
		return value
	}
	if value := p.r.FormValue(name); value != "" {
		return value
	}
	return defaultValue
}

// Int gets an integer parameter from query or form
func (p *Params) Int(name string, defaultValue int) int {
	value := p.String(name, "")
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

// Bool gets a boolean parameter from query or form
func (p *Params) Bool(name string) bool {
	value := strings.ToLower(p.String(name, ""))
	return value == "true" || value == "1" || value == "yes" || value == "on"
}

// Float gets a float parameter from query or form
func (p *Params) Float(name string, defaultValue float64) float64 {
	value := p.String(name, "")
	if value == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return f
}

// Strings gets multiple values for a parameter
func (p *Params) Strings(name string) []string {
	// Check query parameters first
	if values := p.r.URL.Query()[name]; len(values) > 0 {
		return values
	}
	// Check form values
	p.r.ParseForm()
	if values := p.r.Form[name]; len(values) > 0 {
		return values
	}
	return nil
}

// Has checks if a parameter exists
func (p *Params) Has(name string) bool {
	return p.String(name, "") != ""
}

// Pagination helps with paginated results
type Pagination struct {
	Page     int
	PageSize int
	Offset   int
	Limit    int
}

// GetPagination extracts pagination parameters from the request
func GetPagination(r *http.Request, defaultPageSize int) *Pagination {
	p := NewParams(r)
	
	page := p.Int("page", 1)
	if page < 1 {
		page = 1
	}
	
	pageSize := p.Int("page_size", defaultPageSize)
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > 100 {
		pageSize = 100 // Cap at 100 for safety
	}
	
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	}
}

// Sort helps with sorting parameters
type Sort struct {
	Field string
	Order string // "asc" or "desc"
}

// GetSort extracts sort parameters from the request
func GetSort(r *http.Request, defaultField string) *Sort {
	p := NewParams(r)
	
	field := p.String("sort", defaultField)
	order := strings.ToLower(p.String("order", "asc"))
	
	if order != "desc" {
		order = "asc"
	}
	
	return &Sort{
		Field: field,
		Order: order,
	}
}