package authentication

import (
	"context"
	"net/http"
	"github.com/The-Skyscape/devtools/pkg/application"
)

type contextKey string

var (
	sessionKey = contextKey("session")
	userKey    = contextKey("user")
)

func (auth *Controller) ProtectFunc(h http.HandlerFunc, adminOnly bool) http.Handler {
	return auth.Protect(h, adminOnly)
}

func (auth *Controller) Serve(name string, adminOnly bool) http.Handler {
	return auth.App.Serve(name, func(app *application.App, r *http.Request) string {
		if auth.Users.Count() == 0 {
			return "setup.html"
		}

		if user, _, _ := auth.Authenticate(r); user != nil {
			if !adminOnly || user.IsAdmin {
				return ""
			} else {
				return "signin.html"
			}
		}

		return "signin.html"
	})
}

func (auth *Controller) Protect(fn http.Handler, adminOnly bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.setupView != "" && auth.Users.Count() == 0 {
			auth.App.Render(w, r, auth.setupView, nil)
			return
		}
		user, s, _ := auth.Authenticate(r)
		if user == nil || (adminOnly && !user.IsAdmin) {
			auth.App.Render(w, r, auth.signinView, "")
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, sessionKey, s)
		ctx = context.WithValue(ctx, userKey, user)
		fn.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (auth *Controller) Forward(name, to string) http.HandlerFunc {
	view := auth.App.Serve(name, nil)
	return func(w http.ResponseWriter, r *http.Request) {
		if user, _, _ := auth.Authenticate(r); user != nil {
			if htmx := r.Header.Get("HX-Request"); htmx != "" {
				w.Header().Add("Hx-Refresh", "true")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			http.Redirect(w, r, to, http.StatusSeeOther)
			return
		}

		view.ServeHTTP(w, r)
	}
}
