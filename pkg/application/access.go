package application

import "net/http"

type AccessCheck func(*App, http.ResponseWriter, *http.Request) bool

func (app *App) Protect(h http.Handler, accessCheck AccessCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if accessCheck == nil {
			h.ServeHTTP(w, r)
			return
		}

		// If accessCheck returns false, it has already handled the response
		if !accessCheck(app, w, r) {
			return
		}

		h.ServeHTTP(w, r)
	}
}

func (app *App) ProtectFunc(fn http.HandlerFunc, accessLevel AccessCheck) http.HandlerFunc {
	return app.Protect(fn, accessLevel)
}
