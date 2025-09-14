package authentication

import (
	"cmp"
	"log"
	"net/http"
	"time"
)

type Option func(*Controller)

func WithCookie(name string) Option {
	return func(auth *Controller) {
		auth.cookieName = cmp.Or(name, auth.cookieName)
	}
}

func WithSetupView(view, dest string) Option {
	return func(auth *Controller) {
		auth.setupView = view
		auth.setupRedir = dest
	}
}

func WithSignupHandler(fn func(*Controller, *User) http.HandlerFunc) Option {
	return func(auth *Controller) {
		auth.signupFunc = fn
	}
}

func WithSigninHandler(fn func(*Controller, *User) http.HandlerFunc) Option {
	return func(auth *Controller) {
		auth.signinFunc = fn
	}
}

func WithSigninView(view, dest string) Option {
	return func(auth *Controller) {
		auth.signinView = view
		auth.signinRedir = dest
	}
}

func WithSignoutURL(url string) Option {
	return func(d *Controller) {
		if url == "" {
			// Use a sensible default instead of failing
			d.signoutRedir = "/"
			log.Printf("WARNING: empty signout redirect URL provided, using default '/'")
		} else {
			d.signoutRedir = url
		}
	}
}

// WithInactivityTimeout sets the session inactivity timeout
// Sessions expire after this duration of inactivity
func WithInactivityTimeout(timeout time.Duration) Option {
	return func(auth *Controller) {
		auth.inactivityTimeout = timeout
	}
}

// WithAbsoluteTimeout sets the absolute session timeout
// Sessions expire after this duration regardless of activity
func WithAbsoluteTimeout(timeout time.Duration) Option {
	return func(auth *Controller) {
		auth.absoluteTimeout = timeout
	}
}

// WithSessionTimeouts sets both inactivity and absolute timeouts
func WithSessionTimeouts(inactivity, absolute time.Duration) Option {
	return func(auth *Controller) {
		auth.inactivityTimeout = inactivity
		auth.absoluteTimeout = absolute
	}
}
