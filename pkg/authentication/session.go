package authentication

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/golang-jwt/jwt/v5"
)

func (*Session) Table() string { return "sessions" }

type Session struct {
	database.Model
	UserID       string
	LastActivity time.Time // Track last activity for timeout
	ExpiresAt    time.Time // Absolute expiration time
	IPAddress    string    // Track IP for security
	UserAgent    string    // Track user agent for security
}

func (s *Session) Token() (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": s.ID,
		"iat": time.Now().Unix(),
		"exp": s.ExpiresAt.Unix(),
	}).SignedString([]byte(os.Getenv("AUTH_SECRET")))
}

// IsExpired checks if the session has expired due to inactivity or absolute timeout
func (s *Session) IsExpired(inactivityTimeout time.Duration) bool {
	now := time.Now()
	// Check absolute expiration
	if now.After(s.ExpiresAt) {
		return true
	}
	// Check inactivity timeout
	if now.Sub(s.LastActivity) > inactivityTimeout {
		return true
	}
	return false
}

// UpdateActivity updates the last activity time for the session
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
}

// Default session timeouts
const (
	DefaultInactivityTimeout = 30 * time.Minute  // Session expires after 30 minutes of inactivity
	DefaultAbsoluteTimeout   = 24 * time.Hour    // Session expires after 24 hours regardless of activity
)

func (auth *Controller) Authenticate(r *http.Request) (*User, *Session, error) {
	cookie, err := r.Cookie(auth.cookieName)
	if err != nil {
		log.Printf("AUTH: No cookie found with name=%s: %v", auth.cookieName, err)
		return nil, nil, err
	}
	log.Printf("AUTH: Found cookie %s, parsing JWT", auth.cookieName)

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("AUTH_SECRET")), nil
	})

	if err != nil {
		log.Printf("AUTH: JWT parse failed: %v", err)
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("AUTH: Failed to get claims from JWT")
		return nil, nil, err
	}
	log.Printf("AUTH: JWT claims: %+v", claims)

	sessionID, ok := claims["sub"].(string)
	if !ok {
		log.Printf("AUTH: No 'sub' claim in JWT")
		return nil, nil, err
	}
	log.Printf("AUTH: Session ID from JWT: %s", sessionID)

	session, err := auth.Sessions.Get(sessionID)
	if err != nil {
		log.Printf("AUTH: Failed to get session %s: %v", sessionID, err)
		return nil, nil, err
	}
	log.Printf("AUTH: Found session for user %s", session.UserID)

	// Check if session has expired
	inactivityTimeout := auth.GetInactivityTimeout()
	if session.IsExpired(inactivityTimeout) {
		// Delete expired session
		auth.Sessions.Delete(session)
		return nil, nil, errors.New("session expired")
	}

	// Update last activity time
	session.UpdateActivity()
	if err := auth.Sessions.Update(session); err != nil {
		// Log but don't fail authentication
		// This prevents blocking users if DB update fails
	}

	user, err := auth.GetUser(session.UserID)
	return user, session, err
}

// GetInactivityTimeout returns the configured inactivity timeout
func (auth *Controller) GetInactivityTimeout() time.Duration {
	if auth.inactivityTimeout > 0 {
		return auth.inactivityTimeout
	}
	return DefaultInactivityTimeout
}

// CleanupExpiredSessions removes expired sessions from the database
// This should be called periodically (e.g., every hour)
func (auth *Controller) CleanupExpiredSessions() error {
	// Delete sessions that have exceeded absolute timeout
	expiredSessions, err := auth.Sessions.Search("WHERE ExpiresAt < ?", time.Now())
	if err != nil {
		return err
	}

	for _, session := range expiredSessions {
		if err := auth.Sessions.Delete(session); err != nil {
			// Log error but continue cleanup
			continue
		}
	}

	// Delete sessions that have exceeded inactivity timeout
	inactivityTimeout := auth.GetInactivityTimeout()
	inactiveSince := time.Now().Add(-inactivityTimeout)
	inactiveSessions, err := auth.Sessions.Search("WHERE LastActivity < ?", inactiveSince)
	if err != nil {
		return err
	}

	for _, session := range inactiveSessions {
		if err := auth.Sessions.Delete(session); err != nil {
			// Log error but continue cleanup
			continue
		}
	}

	return nil
}

// StartSessionCleanup starts a background goroutine to periodically clean up expired sessions
func (auth *Controller) StartSessionCleanup(interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour // Default to hourly cleanup
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := auth.CleanupExpiredSessions(); err != nil {
				// Log error but continue running
			}
		}
	}()
}
