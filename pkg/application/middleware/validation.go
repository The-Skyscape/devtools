package middleware

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// InputValidation provides comprehensive input validation middleware
type InputValidation struct {
	name              string
	maxBodySize       int64
	maxFieldLength    int
	allowedMethods    []string
	contentTypeChecks map[string]func(*http.Request) error
	skipFunc          SkipFunc
}

// ValidationOption configures the InputValidation middleware
type ValidationOption func(*InputValidation)

// NewInputValidation creates a new input validation middleware
func NewInputValidation(opts ...ValidationOption) *InputValidation {
	v := &InputValidation{
		name:           "input-validation",
		maxBodySize:    10 * 1024 * 1024, // 10MB default
		maxFieldLength: 10000,             // 10k chars per field default
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		contentTypeChecks: map[string]func(*http.Request) error{
			"application/json":                  validateJSON,
			"application/x-www-form-urlencoded": validateFormURLEncoded,
			"multipart/form-data":               validateMultipart,
		},
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Name returns the middleware name
func (v *InputValidation) Name() string {
	return v.name
}

// Handle returns the HTTP handler that performs input validation
func (v *InputValidation) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if we should skip validation for this request
		if v.skipFunc != nil && v.skipFunc(r) {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Validate HTTP method
		if !v.isMethodAllowed(r.Method) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 2. Validate and limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, v.maxBodySize)

		// 3. Validate headers for common attacks
		if err := v.validateHeaders(r); err != nil {
			http.Error(w, "Invalid headers", http.StatusBadRequest)
			return
		}

		// 4. Validate query parameters
		if err := v.validateQueryParams(r); err != nil {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// 5. Validate content type and body
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "DELETE" {
			if err := v.validateContentType(r); err != nil {
				http.Error(w, "Invalid content type or body", http.StatusBadRequest)
				return
			}
		}

		// 6. Sanitize path parameters
		v.sanitizePathParams(r)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// isMethodAllowed checks if the HTTP method is allowed
func (v *InputValidation) isMethodAllowed(method string) bool {
	for _, allowed := range v.allowedMethods {
		if method == allowed {
			return true
		}
	}
	return false
}

// validateHeaders checks for common header-based attacks
func (v *InputValidation) validateHeaders(r *http.Request) error {
	// Check for header injection attempts
	for name, values := range r.Header {
		// Check header name
		if containsControlChars(name) {
			return fmt.Errorf("invalid header name: %s", name)
		}

		// Check header values
		for _, value := range values {
			if containsControlChars(value) {
				return fmt.Errorf("invalid header value for %s", name)
			}

			// Check for excessively long headers
			if len(value) > v.maxFieldLength {
				return fmt.Errorf("header %s too long", name)
			}
		}
	}

	// Check for host header injection
	if host := r.Header.Get("Host"); host != "" {
		if !isValidHost(host) {
			return fmt.Errorf("invalid host header")
		}
	}

	return nil
}

// validateQueryParams validates URL query parameters
func (v *InputValidation) validateQueryParams(r *http.Request) error {
	query := r.URL.Query()

	for key, values := range query {
		// Check parameter name
		if containsSQLInjection(key) || containsXSS(key) {
			return fmt.Errorf("invalid query parameter name: %s", key)
		}

		// Check parameter values
		for _, value := range values {
			// Check length
			if len(value) > v.maxFieldLength {
				return fmt.Errorf("query parameter %s too long", key)
			}

			// Check for SQL injection patterns
			if containsSQLInjection(value) {
				return fmt.Errorf("potential SQL injection in parameter %s", key)
			}

			// Check for XSS patterns
			if containsXSS(value) {
				return fmt.Errorf("potential XSS in parameter %s", key)
			}

			// Check for path traversal
			if containsPathTraversal(value) {
				return fmt.Errorf("potential path traversal in parameter %s", key)
			}
		}
	}

	return nil
}

// validateContentType validates the content type and body
func (v *InputValidation) validateContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return nil // No content type is OK for some requests
	}

	// Parse content type
	mediaType := contentType
	if idx := strings.Index(contentType, ";"); idx != -1 {
		mediaType = strings.TrimSpace(contentType[:idx])
	}

	// Check if we have a validator for this content type
	if validator, ok := v.contentTypeChecks[mediaType]; ok {
		return validator(r)
	}

	// Allow unknown content types but don't validate them
	return nil
}

// sanitizePathParams sanitizes path parameters
func (v *InputValidation) sanitizePathParams(r *http.Request) {
	// Sanitize path values (they're already URL-decoded by Go)
	pathValues := r.PathValue
	if pathValues == nil {
		return
	}

	// Note: In Go 1.22+, path values are accessed via r.PathValue(key)
	// This is a placeholder for when we need to sanitize them
}

// Validation helper functions

func validateJSON(r *http.Request) error {
	// Try to decode JSON to check validity
	var data interface{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict mode

	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	return nil
}

func validateFormURLEncoded(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data: %v", err)
	}

	// Validate form fields
	for key, values := range r.PostForm {
		if containsSQLInjection(key) || containsXSS(key) {
			return fmt.Errorf("invalid form field name: %s", key)
		}

		for _, value := range values {
			if containsSQLInjection(value) {
				return fmt.Errorf("potential SQL injection in field %s", key)
			}
			if containsXSS(value) {
				return fmt.Errorf("potential XSS in field %s", key)
			}
		}
	}

	return nil
}

func validateMultipart(r *http.Request) error {
	// Parse multipart form with size limit
	if err := r.ParseMultipartForm(32 << 20); // 32MB max
	err != nil {
		return fmt.Errorf("invalid multipart form: %v", err)
	}

	// Validate form fields
	if r.MultipartForm != nil {
		for key, values := range r.MultipartForm.Value {
			if containsSQLInjection(key) || containsXSS(key) {
				return fmt.Errorf("invalid form field name: %s", key)
			}

			for _, value := range values {
				if containsSQLInjection(value) {
					return fmt.Errorf("potential SQL injection in field %s", key)
				}
				if containsXSS(value) {
					return fmt.Errorf("potential XSS in field %s", key)
				}
			}
		}

		// Validate file uploads
		for _, files := range r.MultipartForm.File {
			for _, file := range files {
				// Check filename for path traversal
				if containsPathTraversal(file.Filename) {
					return fmt.Errorf("invalid filename: %s", file.Filename)
				}

				// Check file size
				if file.Size > 100*1024*1024 { // 100MB max per file
					return fmt.Errorf("file too large: %s", file.Filename)
				}
			}
		}
	}

	return nil
}

// Security pattern detection

func containsControlChars(s string) bool {
	for _, r := range s {
		if r < 32 || r == 127 { // Control characters
			return true
		}
	}
	return false
}

func isValidHost(host string) bool {
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Basic validation - could be enhanced with proper domain validation
	if len(host) > 253 {
		return false
	}

	// Check for invalid characters
	validHostRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	return validHostRegex.MatchString(host)
}

func containsSQLInjection(s string) bool {
	// Common SQL injection patterns (basic detection)
	sqlPatterns := []string{
		"(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript)",
		"(?i)(;|--|/\\*|\\*/|xp_|sp_|0x)",
		"(?i)('|\")(\\s)*(or|and)(\\s)*('|\"|\\d)",
	}

	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, s); matched {
			return true
		}
	}
	return false
}

func containsXSS(s string) bool {
	// Common XSS patterns (basic detection)
	xssPatterns := []string{
		"<script",
		"javascript:",
		"on\\w+\\s*=", // Event handlers like onclick=
		"<iframe",
		"<embed",
		"<object",
		"expression\\(",
		"vbscript:",
		"data:text/html",
	}

	lower := strings.ToLower(s)
	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, lower); matched {
			return true
		}
	}
	return false
}

func containsPathTraversal(s string) bool {
	// Path traversal patterns
	traversalPatterns := []string{
		"\\.\\./",
		"\\.\\.\\\\",
		"%2e%2e/",
		"%2e%2e\\\\",
		"..;",
		"%00",
	}

	for _, pattern := range traversalPatterns {
		if strings.Contains(strings.ToLower(s), pattern) {
			return true
		}
	}
	return false
}

// SanitizeHTML sanitizes HTML input to prevent XSS
func SanitizeHTML(input string) string {
	// Basic HTML escaping
	return html.EscapeString(input)
}

// SanitizeSQL sanitizes input for SQL queries (use parameterized queries instead!)
func SanitizeSQL(input string) string {
	// This is a last resort - always use parameterized queries
	replacer := strings.NewReplacer(
		"'", "''",
		"\\", "\\\\",
		"\x00", "\\x00",
		"\n", "\\n",
		"\r", "\\r",
		"\x1a", "\\x1a",
	)
	return replacer.Replace(input)
}

// ValidateInt validates and parses an integer with bounds checking
func ValidateInt(value string, min, max int) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %s", value)
	}

	if n < min || n > max {
		return 0, fmt.Errorf("integer out of range [%d, %d]: %d", min, max, n)
	}

	return n, nil
}

// ValidateURL validates a URL
func ValidateURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// Check scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: %s", u.Scheme)
	}

	// Check host
	if u.Host == "" {
		return nil, fmt.Errorf("missing host in URL")
	}

	return u, nil
}

// Option functions

// WithMaxBodySize sets the maximum request body size
func WithMaxBodySize(size int64) ValidationOption {
	return func(v *InputValidation) {
		v.maxBodySize = size
	}
}

// WithMaxFieldLength sets the maximum field length
func WithMaxFieldLength(length int) ValidationOption {
	return func(v *InputValidation) {
		v.maxFieldLength = length
	}
}

// WithAllowedMethods sets the allowed HTTP methods
func WithAllowedMethods(methods []string) ValidationOption {
	return func(v *InputValidation) {
		v.allowedMethods = methods
	}
}

// WithValidationSkipFunc sets a function to determine if validation should be skipped
func WithValidationSkipFunc(fn SkipFunc) ValidationOption {
	return func(v *InputValidation) {
		v.skipFunc = fn
	}
}