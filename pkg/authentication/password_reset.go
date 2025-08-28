package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
	"golang.org/x/crypto/bcrypt"
)

// PasswordResetToken stores password reset tokens
type PasswordResetToken struct {
	application.Model
	
	UserID     string    // User requesting password reset
	Token      string    // Reset token
	Used       bool      // Whether token has been used
	ExpiresAt  time.Time // When the token expires
	IPAddress  string    // IP address of requester for security
	UserAgent  string    // Browser info for security
}

// Table returns the database table name
func (p *PasswordResetToken) Table() string {
	return "password_reset_tokens"
}

// IsExpired checks if the reset token has expired
func (p *PasswordResetToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// IsValid checks if the token is valid for use
func (p *PasswordResetToken) IsValid() bool {
	return !p.Used && !p.IsExpired()
}

// GenerateResetToken creates a secure random reset token
func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreatePasswordResetToken creates a new password reset token
func CreatePasswordResetToken(userID, ipAddress, userAgent string, expirationMinutes int) (*PasswordResetToken, error) {
	token, err := GenerateResetToken()
	if err != nil {
		return nil, err
	}
	
	return &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		Used:      false,
		ExpiresAt: time.Now().Add(time.Duration(expirationMinutes) * time.Minute),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}, nil
}

// ResetPasswordConfig holds configuration for password reset
type ResetPasswordConfig struct {
	TokenExpirationMinutes int
	RequireOldPassword     bool // For logged-in password changes
	MinPasswordLength      int
	RequireUppercase       bool
	RequireLowercase       bool
	RequireNumbers         bool
	RequireSpecialChars    bool
}

// DefaultResetPasswordConfig returns default configuration
func DefaultResetPasswordConfig() *ResetPasswordConfig {
	return &ResetPasswordConfig{
		TokenExpirationMinutes: 60,     // 1 hour
		RequireOldPassword:     true,   // For logged-in users
		MinPasswordLength:      8,
		RequireUppercase:       false,  // Optional by default
		RequireLowercase:       false,
		RequireNumbers:         false,
		RequireSpecialChars:    false,
	}
}

// ValidatePasswordStrength checks if a password meets requirements
func ValidatePasswordStrength(password string, config *ResetPasswordConfig) error {
	if len(password) < config.MinPasswordLength {
		return errors.New("password must be at least " + string(rune(config.MinPasswordLength)) + " characters long")
	}
	
	if config.RequireUppercase {
		hasUpper := false
		for _, r := range password {
			if r >= 'A' && r <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return errors.New("password must contain at least one uppercase letter")
		}
	}
	
	if config.RequireLowercase {
		hasLower := false
		for _, r := range password {
			if r >= 'a' && r <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return errors.New("password must contain at least one lowercase letter")
		}
	}
	
	if config.RequireNumbers {
		hasNumber := false
		for _, r := range password {
			if r >= '0' && r <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return errors.New("password must contain at least one number")
		}
	}
	
	if config.RequireSpecialChars {
		hasSpecial := false
		specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		for _, r := range password {
			for _, s := range specialChars {
				if r == s {
					hasSpecial = true
					break
				}
			}
			if hasSpecial {
				break
			}
		}
		if !hasSpecial {
			return errors.New("password must contain at least one special character")
		}
	}
	
	return nil
}

// HashPassword creates a bcrypt hash of a password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ComparePasswords checks if a password matches a hash
func ComparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}