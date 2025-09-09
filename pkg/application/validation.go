// Package application provides validation helpers and error handling for web applications
package application

import (
	"errors"
	"fmt"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// ErrInvalidEmail indicates an invalid email address
	ErrInvalidEmail = errors.New("invalid email address")
	
	// ErrPasswordTooShort indicates password doesn't meet minimum length
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	
	// ErrPathTraversal indicates an attempt to access files outside allowed directory
	ErrPathTraversal = errors.New("path traversal not allowed")
	
	// ErrInvalidUsername indicates username doesn't meet requirements
	ErrInvalidUsername = errors.New("username must be 3-30 characters, alphanumeric with dashes and underscores")
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,28}[a-zA-Z0-9]$`)
	slugRegex     = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	
	// Additional check for basic structure
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrInvalidEmail
	}
	
	return nil
}

// ValidatePassword checks if a password meets requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// ValidateUsername checks if a username is valid
func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 30 {
		return ErrInvalidUsername
	}
	
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	
	return nil
}

// ValidateSlug checks if a slug is valid (lowercase, hyphenated)
func ValidateSlug(slug string) error {
	if slug == "" {
		return errors.New("slug cannot be empty")
	}
	
	if len(slug) > 100 {
		return errors.New("slug too long")
	}
	
	if !slugRegex.MatchString(slug) {
		return errors.New("slug must be lowercase letters, numbers, and hyphens")
	}
	
	return nil
}

// IsSubPath checks if a path is within the allowed directory
func IsSubPath(basePath, userPath string) bool {
	// Clean and resolve both paths
	base := filepath.Clean(basePath)
	user := filepath.Clean(userPath)
	
	// Make absolute if needed
	if !filepath.IsAbs(user) {
		user = filepath.Join(base, user)
	}
	
	// Check if user path starts with base path
	rel, err := filepath.Rel(base, user)
	if err != nil {
		return false
	}
	
	// Check for path traversal
	return !strings.HasPrefix(rel, "..") && !strings.HasPrefix(rel, "/")
}

// SanitizePath removes dangerous characters from file paths
func SanitizePath(path string) string {
	// Remove null bytes
	path = strings.ReplaceAll(path, "\x00", "")
	
	// Clean the path
	path = filepath.Clean(path)
	
	// Remove leading slashes for relative paths
	path = strings.TrimPrefix(path, "/")
	
	return path
}

// ValidateRequired checks if required fields are present
func ValidateRequired(fields map[string]string) error {
	var missing []string
	
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("required fields missing: %s", strings.Join(missing, ", "))
	}
	
	return nil
}

// ValidateLength checks if a string is within length bounds
func ValidateLength(value string, min, max int) error {
	length := len(value)
	
	if length < min {
		return fmt.Errorf("must be at least %d characters", min)
	}
	
	if max > 0 && length > max {
		return fmt.Errorf("must be at most %d characters", max)
	}
	
	return nil
}

// ValidateEnum checks if a value is in the allowed set
func ValidateEnum(value string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("must be one of: %s", strings.Join(allowed, ", "))
}

// ValidationError represents multiple validation errors
type ValidationError struct {
	Errors map[string]string
}

// Error implements the error interface
func (v ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	
	var parts []string
	for field, msg := range v.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	
	return strings.Join(parts, "; ")
}

// HasErrors checks if there are any validation errors
func (v ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

// Validator helps build validation errors
type Validator struct {
	errors map[string]string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

// AddError adds a validation error for a field
func (v *Validator) AddError(field, message string) {
	v.errors[field] = message
}

// CheckEmail validates an email field
func (v *Validator) CheckEmail(field, email string) {
	if err := ValidateEmail(email); err != nil {
		v.AddError(field, err.Error())
	}
}

// CheckPassword validates a password field
func (v *Validator) CheckPassword(field, password string) {
	if err := ValidatePassword(password); err != nil {
		v.AddError(field, err.Error())
	}
}

// CheckRequired validates a required field
func (v *Validator) CheckRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
	}
}

// CheckLength validates field length
func (v *Validator) CheckLength(field, value string, min, max int) {
	if err := ValidateLength(value, min, max); err != nil {
		v.AddError(field, err.Error())
	}
}

// Result returns the validation result
func (v *Validator) Result() error {
	if len(v.errors) == 0 {
		return nil
	}
	return ValidationError{Errors: v.errors}
}