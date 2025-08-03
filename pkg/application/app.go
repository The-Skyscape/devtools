package application

import (
	"cmp"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

func Serve(views fs.FS, opts ...Option) {
	log.Printf("🚀 Starting Skyscape Application...")
	log.Printf("📱 Visit: http://localhost:%s", cmp.Or(os.Getenv("PORT"), "8080"))

	app := New(views, opts...)
	app.Start()
}

type App struct {
	controllers map[string]Controller
	viewEngine  *template.Template
	hostPrefix  string
	views       []fs.FS
	theme       string
}

func New(views fs.FS, opts ...Option) *App {
	app := App{
		controllers: map[string]Controller{},
		views:       []fs.FS{appViews},
		theme:       "retro",
	}

	if views != nil {
		app.views = append(app.views, views)

		if _, err := fs.Sub(views, "views/public"); err == nil {
			public, _ := fs.Sub(views, "views")
			http.Handle("GET /public/", http.FileServerFS(public))
		}
	}

	for _, opt := range opts {
		if err := opt(&app); err != nil {
			log.Fatal("Failed to setup Congo server:", err)
		}
	}

	return &app
}

// Use returns the controller with the given name
func (app App) Use(name string) Controller {
	return app.controllers[name]
}

// Start runs the application HTTP server and SSL server
func (app *App) Start() error {
	log.Println("Starting Application...")

	app.prepareViews()

	go func() {
		cert := cmp.Or(os.Getenv("CONGO_SSL_FULLCHAIN"), "/root/fullchain.pem")
		if _, err := os.Stat(cert); err != nil {
			log.Println("No SSL Certificate found at:", cert)
			return
		}

		key := cmp.Or(os.Getenv("CONGO_SSL_PRIVKEY"), "/root/privkey.pem")
		if _, err := os.Stat(key); err != nil {
			log.Println("No SSL Key found at:", key)
			return
		}

		if cert != "" && key != "" {
			log.Print("Serving Secure Congo @ https://localhost:443")
			log.Fatal(http.ListenAndServeTLS("0.0.0.0:443", cert, key, nil))
		}
	}()

	addr := "0.0.0.0:" + cmp.Or(os.Getenv("PORT"), "5000")
	log.Print("Serving Unsecure Congo @ http://" + addr)
	return http.ListenAndServe(addr, nil)
}

func (app *App) Server() (string, http.Handler) {
	addr := "0.0.0.0:" + cmp.Or(os.Getenv("PORT"), "5000")
	log.Print("Serving Unsecure Congo @ http://" + addr)
	return addr, nil
}

// Render renders a view with given data to the http writer
func (app *App) Render(w io.Writer, r *http.Request, page string, data any) {
	funcs := template.FuncMap{
		// {{req.URL.Query.Get "search"}}
		"req": func() *http.Request { return r },
		// {{host}}
		"host": func() string { return app.hostPrefix },
		// {{if path_eq "project" .ID "settings"}} ... {{end}}
		"path_eq": func(parts ...string) bool {
			path := fmt.Sprintf("/%s", strings.Join(parts, "/"))
			return r.URL.Path == path
		},
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl.Handle(r) }
	}

	view := app.viewEngine.Lookup(page)
	if view == nil {
		log.Println("view not found", page)
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, "view not found 1", http.StatusNotFound)
			return
		} else {
			fmt.Fprintf(w, "view not found 2")
			os.Exit(1)
		}
	}

	if err := view.Funcs(funcs).Execute(w, data); err != nil {
		log.Print("Error rendering: ", err)
		app.viewEngine.ExecuteTemplate(w, "error-message", err)
	}
}
