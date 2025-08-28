package authentication

import (
	"regexp"
	"strings"
)

var (
	// Email validation regex - RFC 5322 simplified
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	
	// Username validation regex - alphanumeric, underscore, hyphen
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
)

// ValidationConfig holds validation rules
type ValidationConfig struct {
	MinUsernameLength   int
	MaxUsernameLength   int
	MinNameLength       int
	MaxNameLength       int
	AllowedEmailDomains []string // If set, only these domains are allowed
	BlockedEmailDomains []string // These domains are blocked
	RequireUniqueEmail  bool
	RequireUniqueUsername bool
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MinUsernameLength:     3,
		MaxUsernameLength:     30,
		MinNameLength:         2,
		MaxNameLength:         100,
		AllowedEmailDomains:   []string{}, // Allow all by default
		BlockedEmailDomains:   []string{}, // Block none by default
		RequireUniqueEmail:    true,
		RequireUniqueUsername: true,
	}
}

// IsValidEmail checks if an email address is valid
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// IsValidEmailWithConfig checks email with additional configuration
func IsValidEmailWithConfig(email string, config *ValidationConfig) bool {
	if !IsValidEmail(email) {
		return false
	}
	
	email = strings.ToLower(email)
	domain := strings.Split(email, "@")[1]
	
	// Check allowed domains
	if len(config.AllowedEmailDomains) > 0 {
		allowed := false
		for _, allowedDomain := range config.AllowedEmailDomains {
			if domain == strings.ToLower(allowedDomain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	
	// Check blocked domains
	for _, blockedDomain := range config.BlockedEmailDomains {
		if domain == strings.ToLower(blockedDomain) {
			return false
		}
	}
	
	return true
}

// IsValidUsername checks if a username is valid
func IsValidUsername(username string) bool {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	return usernameRegex.MatchString(username)
}

// IsValidUsernameWithConfig checks username with configuration
func IsValidUsernameWithConfig(username string, config *ValidationConfig) bool {
	username = strings.TrimSpace(username)
	if len(username) < config.MinUsernameLength || len(username) > config.MaxUsernameLength {
		return false
	}
	return usernameRegex.MatchString(username)
}

// IsValidName checks if a name is valid
func IsValidName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return false
	}
	// Allow letters, spaces, hyphens, apostrophes
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || 
			(r >= 'A' && r <= 'Z') || 
			r == ' ' || r == '-' || r == '\'') {
			return false
		}
	}
	return true
}

// IsValidNameWithConfig checks name with configuration
func IsValidNameWithConfig(name string, config *ValidationConfig) bool {
	name = strings.TrimSpace(name)
	if len(name) < config.MinNameLength || len(name) > config.MaxNameLength {
		return false
	}
	// Allow letters, spaces, hyphens, apostrophes, and some unicode
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || 
			(r >= 'A' && r <= 'Z') || 
			r == ' ' || r == '-' || r == '\'' || r == '.' ||
			(r >= 0x00C0 && r <= 0x024F)) { // Latin extended characters
			return false
		}
	}
	return true
}

// SanitizeEmail cleans and normalizes an email address
func SanitizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// SanitizeUsername cleans and normalizes a username
func SanitizeUsername(username string) string {
	return strings.TrimSpace(strings.ToLower(username))
}

// SanitizeName cleans a name while preserving case
func SanitizeName(name string) string {
	return strings.TrimSpace(name)
}

// ValidateSignupInput validates all signup form inputs
func ValidateSignupInput(name, email, username, password string, config *ValidationConfig, passwordConfig *ResetPasswordConfig) error {
	if !IsValidNameWithConfig(name, config) {
		return &ValidationError{Field: "name", Message: "Invalid name format"}
	}
	
	if !IsValidEmailWithConfig(email, config) {
		return &ValidationError{Field: "email", Message: "Invalid email address"}
	}
	
	if !IsValidUsernameWithConfig(username, config) {
		return &ValidationError{Field: "username", Message: "Invalid username format"}
	}
	
	if err := ValidatePasswordStrength(password, passwordConfig); err != nil {
		return &ValidationError{Field: "password", Message: err.Error()}
	}
	
	return nil
}

// ValidationError provides field-specific error information
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// IsDisposableEmail checks if an email is from a disposable email service
func IsDisposableEmail(email string) bool {
	// Common disposable email domains (can be extended)
	disposableDomains := []string{
		"guerrillamail.com",
		"mailinator.com",
		"10minutemail.com",
		"tempmail.com",
		"throwaway.email",
		"yopmail.com",
		"maildrop.cc",
		"dispostable.com",
		"temp-mail.org",
		"trashmail.com",
	}
	
	domain := strings.Split(strings.ToLower(email), "@")[1]
	for _, disposable := range disposableDomains {
		if domain == disposable {
			return true
		}
	}
	return false
}