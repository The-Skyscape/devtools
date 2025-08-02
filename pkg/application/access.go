package application

import "net/http"

type AccessCheck func(*App, *http.Request) string

func (app *App) Protect(h http.Handler, accessCheck AccessCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if accessCheck == nil {
			h.ServeHTTP(w, r)
			return
		}

		if page := accessCheck(app, r); page != "" {
			app.Render(w, r, page, nil)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func (app *App) ProtectFunc(fn http.HandlerFunc, accessLevel AccessCheck) http.HandlerFunc {
	return app.Protect(fn, accessLevel)
}
