package authentication

import (
	"testing"
	"time"

)

func TestEmailVerificationTable(t *testing.T) {
	verification := &EmailVerification{}
}

func TestUserVerificationTable(t *testing.T) {
	verification := &UserVerification{}
}

func TestEmailVerificationIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "future expiration",
			expiresAt: time.Now().Add(time.Hour * 2),
			expected:  false,
		},
		{
			name:      "past expiration",
			expiresAt: time.Now().Add(-time.Hour),
			expected:  true,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-time.Second),
			expected:  true,
		},
		{
			name:      "expires in far future",
			expiresAt: time.Now().Add(time.Hour * 24 * 7),
			expected:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verification := &EmailVerification{
				ExpiresAt: tt.expiresAt,
			}
			
			result := verification.IsExpired()
		})
	}
}

func TestEmailVerificationMarkAsUsed(t *testing.T) {
	verification := &EmailVerification{
		Used: false,
	}
	
	beforeMark := time.Now()
	err := verification.MarkAsUsed()
	afterMark := time.Now()
	
}

func TestGenerateVerificationToken(t *testing.T) {
	// Test multiple generations
	tokens := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		token, err := GenerateVerificationToken()
		
		
		// Check uniqueness
		tokens[token] = true
		
		// Check it's valid hex
		for _, char := range token {
			valid := (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')
		}
	}
}

func TestCreateEmailVerification(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		email            string
		expirationHours  int
	}{
		{
			name:            "standard verification",
			userID:          "user-123",
			email:           "user@example.com",
			expirationHours: 24,
		},
		{
			name:            "different user",
			userID:          "user-456",
			email:           "another@example.com",
			expirationHours: 48,
		},
		{
			name:            "short expiration",
			userID:          "user-789",
			email:           "short@example.com",
			expirationHours: 1,
		},
		{
			name:            "long expiration",
			userID:          "user-999",
			email:           "long@example.com",
			expirationHours: 168, // 1 week
		},
	}
	
	var verifications []*EmailVerification
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeCreation := time.Now()
			verification := CreateEmailVerification(tt.userID, tt.email, tt.expirationHours)
			afterCreation := time.Now()
			
			
			// Check expiration time
			expectedExpiry := beforeCreation.Add(time.Duration(tt.expirationHours) * time.Hour)
			
			// Verification should not be expired
			
			verifications = append(verifications, verification)
		})
	}
	
	// All tokens should be unique
	for i, v1 := range verifications {
		for j, v2 := range verifications {
			if i != j {
			}
		}
	}
}

func TestDefaultVerificationConfig(t *testing.T) {
	config := DefaultVerificationConfig()
	
}

func TestEmailVerificationWorkflow(t *testing.T) {
	userID := "test-user"
	email := "test@example.com"
	expirationHours := 24
	
	// 1. Create verification
	verification := CreateEmailVerification(userID, email, expirationHours)
	
	// 2. Simulate verification process
	originalToken := verification.Token
	
	// Token should be valid for lookup
	
	// 3. Mark as used
	err := verification.MarkAsUsed()
	
	// Token should remain the same
	
	// UpdatedAt should be set
}

func TestEmailVerificationExpiry(t *testing.T) {
	userID := "expire-test-user"
	email := "expire@example.com"
	
	// Create verification with very short expiration for testing
	verification := &EmailVerification{
		UserID:    userID,
		Email:     email,
		Token:     "test-token",
		Used:      false,
		ExpiresAt: time.Now().Add(-time.Hour), // Already expired
	}
	
	
	// Create non-expired verification
	nonExpiredVerification := &EmailVerification{
		UserID:    userID,
		Email:     email,
		Token:     "test-token-2",
		Used:      false,
		ExpiresAt: time.Now().Add(time.Hour), // Not expired
	}
	
}

func TestEmailVerificationSecurity(t *testing.T) {
	// Test that verification tokens are cryptographically secure
	
	// Generate many tokens and check for patterns
	tokens := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		token, err := GenerateVerificationToken()
		tokens[i] = token
	}
	
	// Check that no token appears twice
	tokenMap := make(map[string]bool)
	for i, token := range tokens {
		tokenMap[token] = true
	}
	
	// Check that tokens don't have obvious patterns (test first 10 for efficiency)
	for i, token := range tokens[:10] {
		// Should not be all the same character
		firstChar := token[0]
		allSame := true
		for _, char := range token {
			if char != firstChar {
				allSame = false
				break
			}
		}
		
		// Should not be sequential
		isSequential := true
		for j := 1; j < len(token); j++ {
			if token[j] != token[j-1]+1 {
				isSequential = false
				break
			}
		}
	}
}

func TestUserVerificationStates(t *testing.T) {
	userID := "verification-state-user"
	
	// Test initial state
	verification := &UserVerification{
		UserID:        userID,
		EmailVerified: false,
		PhoneVerified: false,
	}
	
	
	// Test email verification
	beforeVerification := time.Now()
	verification.EmailVerified = true
	verification.VerifiedAt = time.Now()
	afterVerification := time.Now()
	
	
	// Phone should still be unverified
}

func TestEmailVerificationMultipleUsers(t *testing.T) {
	users := []struct {
		userID string
		email  string
	}{
		{"user-1", "user1@example.com"},
		{"user-2", "user2@example.com"},
		{"user-3", "user3@example.com"},
		{"user-4", "user4@example.com"},
		{"user-5", "user5@example.com"},
	}
	
	verifications := make([]*EmailVerification, len(users))
	
	// Create verifications for all users
	for i, user := range users {
		verification := CreateEmailVerification(user.userID, user.email, 24)
		verifications[i] = verification
		
	}
	
	// All tokens should be unique
	for i, v1 := range verifications {
		for j, v2 := range verifications {
			if i != j {
			}
		}
	}
	
	// Verify some users
	for i := 0; i < 3; i++ {
		err := verifications[i].MarkAsUsed()
	}
	
	// Others should remain unused
	for i := 3; i < len(verifications); i++ {
	}
}

func TestVerificationConfigCustomization(t *testing.T) {
	customConfig := &VerificationConfig{
		RequireEmailVerification: true,
		TokenExpirationHours:     48,
		ResendCooldownMinutes:    10,
		MaxResendAttempts:        3,
	}
	
}

func TestEmailVerificationEdgeCases(t *testing.T) {
	t.Run("zero expiration hours", func(t *testing.T) {
		verification := CreateEmailVerification("user", "user@example.com", 0)
		// Should expire immediately
	})
	
	t.Run("negative expiration hours", func(t *testing.T) {
		verification := CreateEmailVerification("user", "user@example.com", -1)
		// Should already be expired
	})
	
	t.Run("very large expiration hours", func(t *testing.T) {
		verification := CreateEmailVerification("user", "user@example.com", 8760) // 1 year
		
		// Should expire approximately 1 year from now
		oneYearFromNow := time.Now().Add(time.Hour * 24 * 365)
	})
	
	t.Run("empty user ID", func(t *testing.T) {
		verification := CreateEmailVerification("", "user@example.com", 24)
	})
	
	t.Run("empty email", func(t *testing.T) {
		verification := CreateEmailVerification("user-123", "", 24)
	})
}

func TestEmailVerificationTimestamps(t *testing.T) {
	userVerification := &UserVerification{
		UserID: "timestamp-test-user",
	}
	
	// Initially no timestamps
	
	// Set email verification
	beforeEmail := time.Now()
	userVerification.EmailVerified = true
	userVerification.VerifiedAt = time.Now()
	afterEmail := time.Now()
	
	
	// Set phone verification
	time.Sleep(time.Millisecond) // Ensure different timestamp
	beforePhone := time.Now()
	userVerification.PhoneVerified = true
	userVerification.PhoneVerifiedAt = time.Now()
	afterPhone := time.Now()
	
	
	// Email timestamp should remain unchanged
}

func TestEmailVerificationIntegration(t *testing.T) {
	// Test a complete email verification flow
	userID := "integration-test-user"
	email := "integration@example.com"
	
	// 1. Create verification request
	verification := CreateEmailVerification(userID, email, 24)
	
	originalToken := verification.Token
	
	// 2. Simulate email sent and user clicks link
	// Token would be used to look up verification...
	
	// 3. Process verification
	
	err := verification.MarkAsUsed()
	
	// 4. Create user verification record
	userVerification := &UserVerification{
		UserID:        userID,
		EmailVerified: true,
		VerifiedAt:    time.Now(),
	}
	
	
	// 5. Token should remain the same throughout
}

func TestVerificationTokenCollision(t *testing.T) {
	// Test for extremely unlikely but possible token collisions
	// Generate a reasonable number of tokens to test uniqueness
	const numTokens = 10000
	tokens := make(map[string]bool, numTokens)
	
	for i := 0; i < numTokens; i++ {
		token, err := GenerateVerificationToken()
		
		// Check for collision
		tokens[token] = true
	}
	
	// Verify we generated the expected number of unique tokens
}