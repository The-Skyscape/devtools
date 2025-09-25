package authentication

import (
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
	UserID    string
	ExpiresAt time.Time // Absolute expiration time
}

func (s *Session) Token() (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": s.ID,
		"iat": time.Now().Unix(),
		"exp": s.ExpiresAt.Unix(),
	}).SignedString([]byte(os.Getenv("AUTH_SECRET")))
}

// IsExpired checks if the session has expired due to inactivity or absolute timeout
func (s *Session) IsExpired() bool {
	now := time.Now()
	// Check absolute expiration
	if now.After(s.ExpiresAt) {
		return true
	}
	return false
}

func (auth *Controller) Authenticate(r *http.Request) (*User, *Session, error) {
	cookie, err := r.Cookie(auth.cookieName)
	if err != nil {
		return nil, nil, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("AUTH_SECRET")), nil
	})

	if err != nil {
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, err
	}

	sessionID, ok := claims["sub"].(string)
	if !ok {
		return nil, nil, err
	}

	session, err := auth.Sessions.Get(sessionID)
	if err != nil {
		return nil, nil, err
	}

	user, err := auth.LookupUser(session.UserID)
	return user, session, err
}

// CleanupExpiredSessions removes expired sessions from the database
// This should be called periodically (e.g., every hour)
func (auth *Controller) CleanupExpiredSessions(interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour // Default to hourly cleanup
	}

	for range time.Tick(interval) {
		expired, err := auth.Sessions.Search(`
			WHERE ExpiresAt < ?
		`, time.Now())

		if err != nil {
			log.Println("Failed to cleanup expired sessions:", err)
			continue
		}

		for _, session := range expired {
			if err := auth.Sessions.Delete(session); err != nil {
				// Log error but continue cleanup
				continue
			}
		}
	}
}
