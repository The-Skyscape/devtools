package authentication

import (
	"errors"
	"net/http"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

func (c *Collection) Controller(opts ...Option) *Controller {
	auth := Controller{
		Collection:   c,
		cookieName:   "theskyscape",
		setupView:    "signup.html",
		signinView:   "signin.html",
		signoutRedir: "/",
	}

	for _, opt := range opts {
		opt(&auth)
	}

	return &auth
}

type Controller struct {
	application.BaseController
	*Collection

	// Frontend state
	cookieName string

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
}

func (auth *Controller) Optional(app *application.App, r *http.Request) string {
	return ""
}

func (auth *Controller) Required(app *application.App, r *http.Request) string {
	if auth.Users.Count() == 0 {
		return "signup.html"
	}

	if u, _, err := auth.Authenticate(r); u != nil && err == nil {
		return ""
	}

	return "signin.html"
}

func (auth *Controller) AdminOnly(app *application.App, r *http.Request) string {
	if auth.Users.Count() == 0 {
		return "signup.html"
	}

	if u, _, err := auth.Authenticate(r); u != nil && err == nil && u.IsAdmin {
		return ""
	}

	return "signin.html"
}

func (auth *Controller) Setup(app *application.App) {
	auth.BaseController.Setup(app)
	http.HandleFunc("POST /_auth/signup", auth.HandleSignup)
	http.HandleFunc("POST /_auth/signin", auth.HandleSignin)
	http.HandleFunc("POST /_auth/signout", auth.HandleSignout)
}

func (auth Controller) Handle(r *http.Request) application.Controller {
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
	if user, ok := auth.Context().Value(userKey).(*User); ok {
		return user
	}

	if user, _, err := auth.Authenticate(auth.Request); err == nil {
		return user
	}

	return nil
}

func (auth Controller) HandleSignup(w http.ResponseWriter, r *http.Request) {
	name, handle, email, password := r.FormValue("name"), r.FormValue("handle"), r.FormValue("email"), r.FormValue("password")
	if name == "" || handle == "" || email == "" || password == "" {
		auth.Render(w, r, "error-message", errors.New("missing required fields"))
		return
	}

	user, err := auth.Signup(name, email, handle, password, auth.Users.Count() == 0)
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	session, err := auth.Sessions.Insert(&Session{UserID: user.ID})
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	token, _ := session.Token()
	http.SetCookie(w, &http.Cookie{
		Name:     auth.cookieName,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour * 24 * 4),
		HttpOnly: true,
		Secure:   r.Proto == "https",
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

	user, err := auth.GetUser(handle)
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	if !user.VerifyPassword(password) {
		auth.Render(w, r, "error-message", errors.New("invalid password"))
		return
	}

	session, err := auth.Sessions.Insert(&Session{UserID: user.ID})
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	token, _ := session.Token()
	http.SetCookie(w, &http.Cookie{
		Name:     auth.cookieName,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour * 24 * 4),
		HttpOnly: true,
		Secure:   r.Proto == "https",
	})

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
			Secure:   r.Proto == "https",
		})
	}

	auth.Redirect(w, r, auth.signoutRedir)
}
