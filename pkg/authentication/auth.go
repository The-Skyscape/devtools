package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"time"

	"context"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Auth provides core authentication primitives
// This is a toolkit, not a framework - apps compose what they need
type Auth struct {
	secret string
}

// New creates a new authentication toolkit instance
func New(secret string) *Auth {
	if secret == "" {
		secret = os.Getenv("AUTH_SECRET")
	}
	return &Auth{secret: secret}
}

// HashPassword hashes a password using bcrypt
func (a *Auth) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// VerifyPassword verifies a password against a hash
func (a *Auth) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken creates a JWT token with custom claims
func (a *Auth) GenerateToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

// GenerateSessionToken creates a standard session token
func (a *Auth) GenerateSessionToken(userID string, expiresIn time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": GenerateID(),
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(expiresIn).Unix(),
	}
	return a.GenerateToken(claims)
}

// ValidateToken validates and parses a JWT token
func (a *Auth) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetTokenFromCookie extracts a token from a cookie
func (a *Auth) GetTokenFromCookie(r *http.Request, cookieName string) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetTokenFromHeader extracts a token from Authorization header
func (a *Auth) GetTokenFromHeader(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("no authorization header")
	}

	// Remove "Bearer " prefix if present
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:], nil
	}

	return auth, nil
}

// SetCookie sets an HTTP-only secure cookie
func (a *Auth) SetCookie(w http.ResponseWriter, name, value string, expires time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCookie removes a cookie
func (a *Auth) ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// GenerateID generates a random hex ID
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Middleware creates HTTP middleware that extracts auth info
// Apps can use this as a starting point and customize as needed
func (a *Auth) Middleware(cookieName string, getUserFunc func(userID string) (any, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get token from cookie
			token, err := a.GetTokenFromCookie(r, cookieName)
			if err != nil {
				// Try header as fallback
				token, err = a.GetTokenFromHeader(r)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Validate token
			claims, err := a.ValidateToken(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Get user if function provided
			if getUserFunc != nil {
				if userID, ok := claims["user_id"].(string); ok {
					if user, err := getUserFunc(userID); err == nil {
						// Store in context using standard context values
						ctx := r.Context()
						ctx = context.WithValue(ctx, UserContextKey, user)
						ctx = context.WithValue(ctx, ClaimsContextKey, claims)
						r = r.WithContext(ctx)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Context keys for storing auth data
type contextKey string

var (
	UserContextKey   = contextKey("user")
	ClaimsContextKey = contextKey("claims")
)

// GetUser extracts user from context
func GetUser(r *http.Request) any {
	return r.Context().Value(UserContextKey)
}

// GetClaims extracts claims from context
func GetClaims(r *http.Request) jwt.MapClaims {
	if claims, ok := r.Context().Value(ClaimsContextKey).(jwt.MapClaims); ok {
		return claims
	}
	return nil
}
