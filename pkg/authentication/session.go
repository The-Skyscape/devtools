package authentication

import (
	"net/http"
	"os"
	"github.com/The-Skyscape/devtools/pkg/database"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (*Session) Table() string { return "sessions" }

type Session struct {
	database.Model
	UserID string
}

func (s *Session) Token() (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": s.ID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	}).SignedString([]byte(os.Getenv("AUTH_SECRET")))
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

	user, err := auth.GetUser(session.UserID)
	return user, session, err
}
