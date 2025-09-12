package authentication

import (
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// AccessCheck is a function that can be used with app.Serve or app.ProtectFunc
// Returns true if access is allowed, false otherwise
type AccessCheck func(*application.App, http.ResponseWriter, *http.Request) bool

// RequireAuth creates an access check that requires authentication
// Takes the auth toolkit, cookie name, and a function to get the user
func RequireAuth(auth *Auth, cookieName string, getUserFunc func(string) (any, error)) AccessCheck {
	return func(app *application.App, w http.ResponseWriter, r *http.Request) bool {
		// Try to get token from cookie
		token, err := auth.GetTokenFromCookie(r, cookieName)
		if err != nil {
			// Not authenticated
			return false
		}

		// Validate token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			// Invalid token
			return false
		}

		// Check if user exists
		if getUserFunc != nil {
			if userID, ok := claims["user_id"].(string); ok {
				if _, err := getUserFunc(userID); err != nil {
					// User not found
					return false
				}
			}
		}

		return true
	}
}

// RequireAuthWithRender is like RequireAuth but renders a template on failure
func RequireAuthWithRender(auth *Auth, cookieName string, getUserFunc func(string) (any, error), failTemplate string) AccessCheck {
	return func(app *application.App, w http.ResponseWriter, r *http.Request) bool {
		if !RequireAuth(auth, cookieName, getUserFunc)(app, w, r) {
			app.Render(w, r, failTemplate, nil)
			return false
		}
		return true
	}
}

// RequireFunc creates a custom access check
func RequireFunc(checkFunc func(*http.Request) bool) AccessCheck {
	return func(app *application.App, w http.ResponseWriter, r *http.Request) bool {
		return checkFunc(r)
	}
}

// PublicAccess always allows access
func PublicAccess() AccessCheck {
	return func(app *application.App, w http.ResponseWriter, r *http.Request) bool {
		return true
	}
}

// CombineChecks combines multiple access checks with AND logic
func CombineChecks(checks ...AccessCheck) AccessCheck {
	return func(app *application.App, w http.ResponseWriter, r *http.Request) bool {
		for _, check := range checks {
			if !check(app, w, r) {
				return false
			}
		}
		return true
	}
}
