// Package authentication implements user authentication and access control.
package authentication

import (
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// Optional checks if any users exist in the system. If no users exist,
// it renders the setup view and returns false. Otherwise returns true.
// This can be used to allow both authenticated and unauthenticated access.
func (auth *Controller) Optional(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	if auth.Users.Count("") == 0 {
		app.Render(w, r, auth.setupView, nil)
		return false
	}

	return true
}

// Required ensures a user is authenticated and verified. It returns true only if
// a verified user is authenticated. Otherwise, it renders the appropriate view
// (setup, verify, or signin) and returns false.
func (auth *Controller) Required(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	if auth.Users.Count("") == 0 {
		app.Render(w, r, auth.setupView, nil)
		return false
	}

	user, _, err := auth.Authenticate(r)
	if user == nil || err != nil {
		app.Render(w, r, auth.signinView, nil)
		return false
	}

	if !user.Verified && auth.verifyView != "" {
		app.Render(w, r, auth.verifyView, nil)
		return false
	}

	return true
}

// AdminOnly ensures a user is authenticated and has admin privileges.
// It returns true only if an admin user is authenticated. Otherwise,
// it renders the appropriate view (setup or signin) and returns false.
func (auth *Controller) AdminOnly(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	if auth.Users.Count("") == 0 {
		app.Render(w, r, auth.setupView, nil)
		return false
	}

	user, _, err := auth.Authenticate(r)
	if user == nil || err != nil || !user.IsAdmin {
		app.Render(w, r, auth.signinView, nil)
		return false
	}

	return true
}
