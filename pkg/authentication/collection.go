package authentication

import (
	"fmt"
	"time"

	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func Manage(db *database.DynamicDB) *Collection {
	repo := &Collection{
		db:                  db,
		Users:               database.Manage(db, new(User)),
		Sessions:            database.Manage(db, new(Session)),
		EmailVerifications:  database.Manage(db, new(EmailVerification)),
		PasswordResetTokens: database.Manage(db, new(PasswordResetToken)),
	}

	db.Query(`
		CREATE UNIQUE INDEX IF NOT EXISTS unique_handle ON users (handle);
		CREATE UNIQUE INDEX IF NOT EXISTS unique_email ON users (email);
	`).Exec()

	return repo
}

type Collection struct {
	db                  *database.DynamicDB
	Users               *database.Collection[*User]
	Sessions            *database.Collection[*Session]
	EmailVerifications  *database.Collection[*EmailVerification]
	PasswordResetTokens *database.Collection[*PasswordResetToken]
}

func (c *Collection) LookupUser(ident string) (*User, error) {
	return database.Cursor(c.db, new(User), `

		WHERE ID = $1 OR Email = $1 OR Handle = $1
	
	`, ident).One()
}

func (c *Collection) Signup(name, email, handle, password string, isAdmin bool) (*User, error) {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return c.Users.Insert(&User{
		Avatar:   fmt.Sprintf("https://robots.skysca.pe/%s?set=set4", email),
		Name:     name,
		Email:    email,
		Handle:   handle,
		PassHash: passhash,
		IsAdmin:  isAdmin,
		Verified: false,
	})
}

func (c *Collection) Signin(ident string, password string) (user *User, err error) {
	if user, err = c.LookupUser(ident); err != nil {
		return nil, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	return user, errors.Wrap(err, "failed to verify password")
}

// ResetPassword generates and stores a password reset token for a user
func (c *Collection) ResetPassword(user *User) (*PasswordResetToken, error) {
	if user == nil {
		return nil, errors.New("user not found")
	}

	return c.PasswordResetTokens.Insert(&PasswordResetToken{
		UserID:    user.ID,
		Token:     database.RandomString(32),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
}

// NewEmailVerification generates and stores an email verification token for a user
func (c *Collection) NewEmailVerification(user *User) (*EmailVerification, error) {
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check for existing unused verification
	existing, err := c.EmailVerifications.First(`
		WHERE UserID = ? AND Used = false
	`, user.ID)

	if err != nil {
		existing.Token = database.RandomString(32)
		existing.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
		return existing, c.EmailVerifications.Update(existing)
	}

	return c.EmailVerifications.Insert(&EmailVerification{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     database.RandomString(32),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
}

// VerifyToken verifies an email using the provided token
func (c *Collection) VerifyToken(token string) error {
	if token == "" {
		return errors.New("invalid verification token")
	}

	verification, _ := c.EmailVerifications.First(`
		WHERE Token = ? AND Used = false
	`, token)

	if verification == nil {
		return errors.New("invalid or expired verification link")
	}

	if verification.IsExpired() {
		return errors.New("verification link has expired")
	}

	// Mark user as verified
	user, err := c.Users.Get(verification.UserID)
	if err != nil {
		return err
	}
	if user != nil {
		user.Verified = true
		c.Users.Update(user)
	}

	// Mark token as used
	verification.Used = true
	c.EmailVerifications.Update(verification)

	return nil
}

// ConsumeToken resets a user's password using the provided token
func (c *Collection) ConsumeToken(token, newPassword string) error {
	if token == "" || newPassword == "" {
		return errors.New("invalid token or password")
	}

	// Find the reset token
	resetToken, _ := c.PasswordResetTokens.First("WHERE Token = ? AND Used = 0", token)
	if resetToken == nil {
		return errors.New("invalid or expired reset token")
	}

	if resetToken.IsExpired() {
		return errors.New("reset token has expired")
	}

	// Get the user
	user, err := c.Users.Get(resetToken.UserID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Hash the new password
	passhash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user's password
	user.PassHash = passhash
	err = c.Users.Update(user)
	if err != nil {
		return err
	}

	// Mark token as used
	resetToken.Used = true
	c.PasswordResetTokens.Update(resetToken)

	return nil
}

// ValidateResetToken checks if a reset token is valid without using it
func (c *Collection) ValidateResetToken(token string) (*PasswordResetToken, error) {
	if token == "" {
		return nil, errors.New("invalid reset token")
	}

	resetToken, _ := c.PasswordResetTokens.First("WHERE Token = ? AND Used = 0", token)
	if resetToken == nil {
		return nil, errors.New("invalid or expired reset token")
	}

	if resetToken.IsExpired() {
		return nil, errors.New("reset token has expired")
	}

	return resetToken, nil
}
