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
	
	"github.com/The-Skyscape/devtools/pkg/application/builtins"
)

func Serve(views fs.FS, opts ...Option) {
	log.Printf("🚀 Starting Skyscape Application...")
	log.Printf("📱 Visit: http://localhost:%s", cmp.Or(os.Getenv("PORT"), "8080"))

	app := New(views, opts...)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

type App struct {
	controllers map[string]Controller
	viewEngine  *template.Template
	hostPrefix  string
	views       []fs.FS
	theme       string
	middlewares []Middleware
}

type Middleware interface {
	Handle(http.Handler) http.Handler
}

func New(views fs.FS, opts ...Option) *App {
	app := App{
		controllers: map[string]Controller{},
		views:       []fs.FS{appViews},
		theme:       "retro",
		middlewares: []Middleware{},
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
			log.Fatal("Failed to setup Application:", err)
		}
	}

	return &app
}

// Use returns the controller with the given name
func (app App) Use(name string) Controller {
	return app.controllers[name]
}

// Server prepares the application and returns the address and handler
func (app *App) Server() (string, http.Handler) {
	log.Println("Preparing Application...")

	app.prepareViews()

	// Build middleware chain
	var handler http.Handler = http.DefaultServeMux
	for i := len(app.middlewares) - 1; i >= 0; i-- {
		handler = app.middlewares[i].Handle(handler)
	}

	addr := "0.0.0.0:" + cmp.Or(os.Getenv("PORT"), "5000")
	return addr, handler
}

// Start runs the application HTTP server and SSL server
func (app *App) Start() error {
	addr, handler := app.Server()

	go func() {
		cert := cmp.Or(os.Getenv("SKYSCAPE_SSL_FULLCHAIN"), "/root/fullchain.pem")
		if _, err := os.Stat(cert); err != nil {
			log.Println("No SSL Certificate found at:", cert)
			return
		}

		key := cmp.Or(os.Getenv("SKYSCAPE_SSL_PRIVKEY"), "/root/privkey.pem")
		if _, err := os.Stat(key); err != nil {
			log.Println("No SSL Key found at:", key)
			return
		}

		if cert != "" && key != "" {
			log.Print("Serving Secure Application @ https://localhost:443")
			log.Fatal(http.ListenAndServeTLS("0.0.0.0:443", cert, key, handler))
		}
	}()

	log.Print("Serving Application @ http://" + addr)
	return http.ListenAndServe(addr, handler)
}

// SetTheme updates the application theme
func (app *App) SetTheme(theme string) {
	app.theme = theme
}

// Render renders a view with given data to the http writer
func (app *App) Render(w io.Writer, r *http.Request, page string, data any) {
	// Create a copy of built-in functions to avoid modifying the shared map
	funcs := make(template.FuncMap)
	for k, v := range builtins.FuncMap {
		funcs[k] = v
	}

	// Add runtime-specific functions
	funcs["req"] = func() *http.Request { return r }
	funcs["host"] = func() string { return app.hostPrefix }
	funcs["path_eq"] = func(parts ...string) bool {
		path := fmt.Sprintf("/%s", strings.Join(parts, "/"))
		return r.URL.Path == path
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl.Handle(r) }
	}

	view := app.viewEngine.Lookup(page)
	if view == nil {
		log.Printf("Template not found: %s", page)
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, fmt.Sprintf("Template not found: %s", page), http.StatusNotFound)
			return
		} else {
			fmt.Fprintf(w, "Template not found in non-HTTP context: %s", page)
			os.Exit(1)
		}
	}

	if err := view.Funcs(funcs).Execute(w, data); err != nil {
		log.Print("Error rendering: ", err)
		app.viewEngine.ExecuteTemplate(w, "error-message", err)
	}
}
