package application

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Params provides convenient access to request parameters
type Params struct {
	r         *http.Request
	multipart bool
	parsed    bool
}

// NewParams creates a parameter helper for the request
func NewParams(r *http.Request) *Params {
	return &Params{r: r, multipart: false}
}

// NewMultipartParams creates a parameter helper for multipart form requests
func NewMultipartParams(r *http.Request) *Params {
	return &Params{r: r, multipart: true}
}

// ensureParsed ensures the form/multipart data has been parsed
func (p *Params) ensureParsed() error {
	if p.parsed {
		return nil
	}
	
	var err error
	if p.multipart {
		// Parse multipart form with 32MB max memory
		err = p.r.ParseMultipartForm(32 << 20)
	} else {
		err = p.r.ParseForm()
	}
	
	if err != nil {
		return err
	}
	p.parsed = true
	return nil
}

// String gets a string parameter from query or form
func (p *Params) String(name string, defaultValue string) string {
	if value := p.r.URL.Query().Get(name); value != "" {
		return value
	}
	
	// Ensure form is parsed
	p.ensureParsed()
	
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
	
	// Ensure form is parsed
	p.ensureParsed()
	
	// Check form values
	if values := p.r.Form[name]; len(values) > 0 {
		return values
	}
	return nil
}

// Has checks if a parameter exists
func (p *Params) Has(name string) bool {
	return p.String(name, "") != ""
}

// FileHeader represents an uploaded file
type FileHeader struct {
	*multipart.FileHeader
}

// File gets a single uploaded file by field name
func (p *Params) File(name string) (*FileHeader, error) {
	if !p.multipart {
		return nil, nil
	}
	
	if err := p.ensureParsed(); err != nil {
		return nil, err
	}
	
	if p.r.MultipartForm == nil || p.r.MultipartForm.File == nil {
		return nil, nil
	}
	
	files := p.r.MultipartForm.File[name]
	if len(files) == 0 {
		return nil, nil
	}
	
	return &FileHeader{files[0]}, nil
}

// Files gets multiple uploaded files by field name
func (p *Params) Files(name string) ([]*FileHeader, error) {
	if !p.multipart {
		return nil, nil
	}
	
	if err := p.ensureParsed(); err != nil {
		return nil, err
	}
	
	if p.r.MultipartForm == nil || p.r.MultipartForm.File == nil {
		return nil, nil
	}
	
	files := p.r.MultipartForm.File[name]
	if len(files) == 0 {
		return nil, nil
	}
	
	result := make([]*FileHeader, len(files))
	for i, f := range files {
		result[i] = &FileHeader{f}
	}
	return result, nil
}

// SaveFile saves an uploaded file to the specified path
func (p *Params) SaveFile(name string, destPath string) error {
	file, err := p.File(name)
	if err != nil {
		return err
	}
	if file == nil {
		return nil
	}
	
	return file.Save(destPath)
}

// Save saves the uploaded file to the specified path
func (f *FileHeader) Save(destPath string) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	
	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	
	// Copy file contents
	_, err = io.Copy(dst, src)
	return err
}

// ReadAll reads the entire file contents into memory
func (f *FileHeader) ReadAll() ([]byte, error) {
	src, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	
	return io.ReadAll(src)
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