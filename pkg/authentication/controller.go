package authentication

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

func (c *Collection) Controller(opts ...Option) *Controller {
	auth := Controller{
		Collection:       c,
		cookieName:       "theskyscape",
		setupView:        "signup.html",
		signinView:       "signin.html",
		signoutRedir:     "/",
		verificationView: "email-verification-required.html",
	}

	for _, opt := range opts {
		opt(&auth)
	}

	return &auth
}

// isSecureRequest checks if the request is over HTTPS
// It checks multiple indicators since requests may come through proxies
func isSecureRequest(r *http.Request) bool {
	// Check X-Forwarded-Proto header (most common with proxies/load balancers)
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	// Check if TLS is present (direct HTTPS connection)
	if r.TLS != nil {
		return true
	}
	// Check URL scheme
	if r.URL != nil && r.URL.Scheme == "https" {
		return true
	}
	// Fallback to Proto field (less reliable)
	return r.Proto == "https"
}

type Controller struct {
	application.Controller
	*Collection

	// Frontend state
	cookieName string

	// Session management
	inactivityTimeout time.Duration
	absoluteTimeout   time.Duration

	// Setup functions
	setupView  string
	setupRedir string
	signupFunc func(*Controller, *User) http.HandlerFunc

	// Signin functions
	signinView  string
	signinRedir string
	signinFunc  func(*Controller, *User) http.HandlerFunc

	// Signout functions
	signoutRedir string

	// Email verification
	requireVerification bool
	verificationView    string
}

func (auth *Controller) Optional(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	// Optional always allows access
	return true
}

func (auth *Controller) Required(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	if auth.Users.Count("") == 0 {
		app.Render(w, r, "signup.html", nil)
		return false
	}

	if u, _, err := auth.Authenticate(r); u != nil && err == nil {
		// Check if verification is required
		if auth.requireVerification && !u.Verified {
			app.Render(w, r, auth.verificationView, nil)
			return false
		}
		return true
	}

	app.Render(w, r, "signin.html", nil)
	return false
}

func (auth *Controller) AdminOnly(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	if auth.Users.Count("") == 0 {
		app.Render(w, r, "signup.html", nil)
		return false
	}

	if u, _, err := auth.Authenticate(r); u != nil && err == nil && u.IsAdmin {
		return true
	}

	app.Render(w, r, "signin.html", nil)
	return false
}

func (auth *Controller) Setup(app *application.App) {
	auth.Controller.Setup(app)
	http.HandleFunc("POST /_auth/signup", auth.HandleSignup)
	http.HandleFunc("POST /_auth/signin", auth.HandleSignin)
	http.HandleFunc("POST /_auth/signout", auth.HandleSignout)
}

func (auth Controller) Handle(r *http.Request) application.Handler {
	auth.Request = r
	return &auth
}

func (auth *Controller) CurrentSession() *Session {
	if s, ok := auth.Context().Value(sessionKey).(*Session); ok {
		return s
	}

	if _, s, err := auth.Authenticate(auth.Request); err == nil {
		return s
	}

	return nil
}

func (auth *Controller) CurrentUser() *User {
	log.Printf("AUTH: CurrentUser called, Request=%v", auth.Request != nil)
	if user, ok := auth.Context().Value(userKey).(*User); ok {
		log.Printf("AUTH: CurrentUser found in context: %s", user.Email)
		return user
	}

	if auth.Request == nil {
		log.Printf("AUTH: CurrentUser failed - Request is nil")
		return nil
	}

	if user, _, err := auth.Authenticate(auth.Request); err == nil {
		log.Printf("AUTH: CurrentUser authenticated: %s", user.Email)
		return user
	} else {
		log.Printf("AUTH: CurrentUser authentication failed: %v", err)
	}

	return nil
}

func (auth *Controller) IsAuthenticated() bool {
	return auth.CurrentUser() != nil
}

func (auth Controller) HandleSignup(w http.ResponseWriter, r *http.Request) {
	name, handle, email, password := r.FormValue("name"), r.FormValue("handle"), r.FormValue("email"), r.FormValue("password")
	if name == "" || handle == "" || email == "" || password == "" {
		auth.Render(w, r, "error-message", errors.New("missing required fields"))
		return
	}

	user, err := auth.Signup(name, email, handle, password, auth.Users.Count("") == 0)
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	// Get configured timeouts or use defaults
	absoluteTimeout := auth.absoluteTimeout
	if absoluteTimeout == 0 {
		absoluteTimeout = DefaultAbsoluteTimeout
	}

	session, err := auth.Sessions.Insert(&Session{
		UserID:       user.ID,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(absoluteTimeout),
		IPAddress:    getClientIP(r),
		UserAgent:    r.Header.Get("User-Agent"),
	})
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	token, _ := session.Token()
	http.SetCookie(w, &http.Cookie{
		Name:     auth.cookieName,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
	})

	if auth.signupFunc != nil {
		auth.signupFunc(&auth, user)
		return
	}

	if auth.setupRedir != "" {
		http.Redirect(w, r, auth.setupRedir, http.StatusSeeOther)
		return
	}

	auth.Refresh(w, r)
}

func (auth Controller) HandleSignin(w http.ResponseWriter, r *http.Request) {
	handle, password := r.FormValue("handle"), r.FormValue("password")
	log.Printf("AUTH: Signin attempt for handle=%s from IP=%s", handle, getClientIP(r))

	user, err := auth.GetUser(handle)
	if err != nil {
		log.Printf("AUTH: Signin failed - user not found: %s", handle)
		auth.Render(w, r, "error-message", err)
		return
	}

	if !user.VerifyPassword(password) {
		log.Printf("AUTH: Signin failed - invalid password for: %s", handle)
		auth.Render(w, r, "error-message", errors.New("invalid password"))
		return
	}
	log.Printf("AUTH: Password verified for user: %s (ID=%s)", user.Email, user.ID)

	// Get configured timeouts or use defaults
	absoluteTimeout := auth.absoluteTimeout
	if absoluteTimeout == 0 {
		absoluteTimeout = DefaultAbsoluteTimeout
	}

	session, err := auth.Sessions.Insert(&Session{
		UserID:       user.ID,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(absoluteTimeout),
		IPAddress:    getClientIP(r),
		UserAgent:    r.Header.Get("User-Agent"),
	})
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	token, _ := session.Token()
	secure := isSecureRequest(r)
	log.Printf("AUTH: Setting signin cookie: name=%s secure=%v expires=%v proto=%s tls=%v forwarded=%s",
		auth.cookieName, secure, session.ExpiresAt, r.Proto, r.TLS != nil, r.Header.Get("X-Forwarded-Proto"))
	http.SetCookie(w, &http.Cookie{
		Name:     auth.cookieName,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   secure,
	})
	log.Printf("AUTH: Signin successful for %s, redirecting", user.Email)

	if auth.signinFunc != nil {
		auth.signinFunc(&auth, user)
		return
	}

	if auth.signinRedir != "" {
		auth.Redirect(w, r, auth.signinRedir)
		return
	}

	auth.Refresh(w, r)
}

func (auth Controller) HandleSignout(w http.ResponseWriter, r *http.Request) {
	if _, s, _ := auth.Authenticate(r); s != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     auth.cookieName,
			Value:    "",
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(-1),
			HttpOnly: true,
			Secure:   isSecureRequest(r),
		})
	}

	auth.Redirect(w, r, auth.signoutRedir)
}

// getClientIP extracts the client's IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP if multiple are present
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
