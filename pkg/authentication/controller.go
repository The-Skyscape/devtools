package authentication

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
	"golang.org/x/crypto/bcrypt"
)

func (col *Collection) Controller(opts ...Option) *Controller {
	auth := &Controller{
		Collection:   col,
		cookieName:   "theskyscape",
		cookieTTL:    time.Hour * 24 * 30, // 30 days
		setupView:    "signup.html",
		signinView:   "signin.html",
		signoutRedir: "/",
	}

	for _, opt := range opts {
		opt(auth)
	}

	return auth
}

type Controller struct {
	application.Controller
	*Collection

	// Frontend state
	cookieName string
	cookieTTL  time.Duration

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
	verifyView string
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
	_, session, err := auth.Authenticate(auth.Request)
	if err != nil {
		return nil
	}

	return session
}

func (auth *Controller) CurrentUser() *User {
	user, _, err := auth.Authenticate(auth.Request)
	if err != nil {
		return nil
	}

	return user
}

func (auth *Controller) IsAuthenticated() bool {
	return auth.CurrentUser() != nil
}

func (auth *Controller) HandleSignup(w http.ResponseWriter, r *http.Request) {
	name, handle, email, password := r.FormValue("name"), r.FormValue("handle"), r.FormValue("email"), r.FormValue("password")
	if name == "" || handle == "" || email == "" || password == "" {
		auth.Render(w, r, "error-message.html", errors.New("missing required fields"))
		return
	}

	user, err := auth.Signup(name, email, handle, password, auth.Users.Count("") == 0)
	if err != nil {
		auth.Render(w, r, "error-message.html", err)
		return
	}

	session, err := auth.Sessions.Insert(&Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(auth.cookieTTL),
	})

	if err != nil {
		auth.Render(w, r, "error-message.html", err)
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
		auth.signupFunc(auth, user).ServeHTTP(w, r)
		return
	}

	if auth.setupRedir != "" {
		http.Redirect(w, r, auth.setupRedir, http.StatusSeeOther)
		return
	}

	auth.Refresh(w, r)
}

func (auth *Controller) HandleSignin(w http.ResponseWriter, r *http.Request) {
	handle, password := r.FormValue("handle"), r.FormValue("password")

	user, err := auth.LookupUser(handle)
	if err != nil {
		auth.Render(w, r, "error-message.html", err)
		return
	}

	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		auth.Render(w, r, "error-message.html", errors.New("invalid password"))
		return
	}

	session, err := auth.Sessions.Insert(&Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(auth.cookieTTL),
	})
	if err != nil {
		auth.Render(w, r, "error-message.html", err)
		return
	}

	token, _ := session.Token()
	secure := isSecureRequest(r)
	http.SetCookie(w, &http.Cookie{
		Name:     auth.cookieName,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   secure,
	})

	if auth.signinFunc != nil {
		auth.signinFunc(auth, user).ServeHTTP(w, r)
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
			MaxAge:   -1,
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
