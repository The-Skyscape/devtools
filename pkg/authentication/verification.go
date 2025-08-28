package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// EmailVerification tracks email verification tokens for users
type EmailVerification struct {
	application.Model
	
	UserID     string    // User ID who needs to verify
	Email      string    // Email address to verify
	Token      string    // Verification token
	Used       bool      // Whether token has been used
	ExpiresAt  time.Time // When the token expires
}

// Table returns the database table name
func (e *EmailVerification) Table() string {
	return "email_verifications"
}

// IsExpired checks if the verification token has expired
func (e *EmailVerification) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// MarkAsUsed marks the token as used
func (e *EmailVerification) MarkAsUsed() error {
	e.Used = true
	e.UpdatedAt = time.Now()
	// This would need to be saved to database by the caller
	return nil
}

// GenerateVerificationToken creates a secure random token
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// UserVerification tracks whether a user has verified their email
type UserVerification struct {
	application.Model
	
	UserID         string    // User ID
	EmailVerified  bool      // Whether email is verified
	VerifiedAt     time.Time // When email was verified
	PhoneVerified  bool      // Whether phone is verified (future use)
	PhoneVerifiedAt time.Time // When phone was verified (future use)
}

// Table returns the database table name
func (u *UserVerification) Table() string {
	return "user_verifications"
}

// CreateEmailVerification creates a new verification token for a user
func CreateEmailVerification(userID, email string, expirationHours int) *EmailVerification {
	token, _ := GenerateVerificationToken()
	
	return &EmailVerification{
		UserID:    userID,
		Email:     email,
		Token:     token,
		Used:      false,
		ExpiresAt: time.Now().Add(time.Duration(expirationHours) * time.Hour),
	}
}

// VerificationConfig holds configuration for email verification
type VerificationConfig struct {
	RequireEmailVerification bool
	TokenExpirationHours     int
	ResendCooldownMinutes    int
	MaxResendAttempts        int
}

// DefaultVerificationConfig returns default configuration
func DefaultVerificationConfig() *VerificationConfig {
	return &VerificationConfig{
		RequireEmailVerification: false, // Off by default for backwards compatibility
		TokenExpirationHours:     24,
		ResendCooldownMinutes:    5,
		MaxResendAttempts:        5,
	}
}