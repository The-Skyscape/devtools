package application

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//go:embed all:views
var appViews embed.FS

type View struct {
	app         *App
	name        string
	accessCheck AccessCheck
}

func (app *App) Serve(name string, accessCheck AccessCheck) *View {
	return &View{app: app, name: name, accessCheck: accessCheck}
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data any) {
	v.app.Render(w, r, v.name, data)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if v.accessCheck == nil {
		v.app.Render(w, r, v.name, nil)
		return
	}

	if page := v.accessCheck(v.app, r); page != "" {
		v.app.Render(w, r, page, nil)
		return
	}

	v.app.Render(w, r, v.name, nil)
}

func (app *App) prepareViews() {
	funcs := template.FuncMap{
		"req":     func() *http.Request { return nil },
		"host":    func() string { return app.hostPrefix },
		"path":    func(parts ...string) string { return fmt.Sprintf("/%s", strings.Join(parts, "/")) },
		"theme":   func() string { return app.theme },
		"title":   func(title string) string { return strings.ReplaceAll(title, "_", " ") },
		"prefix":  func(s, prefix string) bool { return strings.HasPrefix(s, prefix) },
		"path_eq": func(parts ...string) bool { return false },
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl }
	}

	if app.viewEngine == nil {
		app.viewEngine = template.New("")
	}

	app.viewEngine = app.viewEngine.Funcs(funcs)
	for _, source := range app.views {
		if tmpl, err := app.viewEngine.ParseFS(source, "views/*.html"); err == nil {
			app.viewEngine = tmpl
		} else {
			log.Fatal("Failed to parse root views", err)
		}

		if tmpl, err := app.viewEngine.ParseFS(source, "views/**/*.html"); err == nil {
			app.viewEngine = tmpl
		}

		if tmpl, err := app.viewEngine.ParseFS(source, "views/**/**/*.html"); err == nil {
			app.viewEngine = tmpl
		}
	}
}
