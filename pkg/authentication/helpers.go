package authentication

import (
	"net/http"
)

// Simple helper methods that apps can use directly
// These avoid complex factory patterns and follow idiomatic Go

// GetAuthenticatedUser extracts the authenticated user from the request
// Returns nil if not authenticated - safe to use in templates
func (c *Controller) GetAuthenticatedUser(r *http.Request) *User {
	user, _, _ := c.Authenticate(r)
	return user
}

// GetAuthenticatedSession extracts the session from the request
// Returns nil if not authenticated
func (c *Controller) GetAuthenticatedSession(r *http.Request) *Session {
	_, session, _ := c.Authenticate(r)
	return session
}

// IsAdmin checks if the authenticated user is an admin
func (c *Controller) IsAdmin(r *http.Request) bool {
	user := c.GetAuthenticatedUser(r)
	return user != nil && user.IsAdmin
}

// RequireUser returns the authenticated user or nil
// Simpler than Authenticate for common use cases
func (c *Controller) RequireUser(r *http.Request) *User {
	return c.GetAuthenticatedUser(r)
}

// Login creates a session for the user and sets the cookie
func (c *Controller) Login(w http.ResponseWriter, r *http.Request, user *User) error {
	session, err := c.Sessions.Insert(&Session{UserID: user.ID})
	if err != nil {
		return err
	}
	
	c.SetSessionCookie(w, r, session)
	return nil
}

// Logout destroys the session and clears the cookie
func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
	})
}

// SetSessionCookie sets the session cookie on the response
func (c *Controller) SetSessionCookie(w http.ResponseWriter, r *http.Request, session *Session) {
	token, _ := session.Token()
	http.SetCookie(w, &http.Cookie{
		Name:     c.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	})
}