package application

import (
	"cmp"
	"html/template"
	"io/fs"
	"log"
)

// Option is a function that configures an Application
type Option func(*App) error

// WithFunc adds a template function to the application
func WithFunc(name string, fn any) Option {
	return func(app *App) error {
		app.viewEngine.Funcs(template.FuncMap{name: fn})
		return nil
	}
}

// WithController adds a controller to the application
func WithController(name string, ctrl Controller) Option {
	return func(app *App) error {
		return app.WithController(name, ctrl)
	}
}

// WithController adds a controller to the application
func (app *App) WithController(name string, controller Controller) error {
	if _, ok := app.controllers[name]; !ok {
		app.controllers[name] = controller
		controller.Setup(app)
	} else {
		log.Fatal(name, "already registered controller")
	}
	return nil
}

// WithViews adds views directory to application
func WithViews(views fs.FS) Option {
	return func(app *App) error {
		return app.WithViews(views)
	}
}

// WithViews adds views directory to application
func (app *App) WithViews(source fs.FS) error {
	app.views = append(app.views, source)
	return nil
}

// WithHostPrefix sets the host prefix for the views
func WithHostPrefix(prefix string) Option {
	return func(app *App) error {
		app.hostPrefix = prefix
		return nil
	}
}

// WithDaisyTheme sets the theme for the views
func WithDaisyTheme(theme string) Option {
	return func(app *App) error {
		app.theme = cmp.Or(theme, app.theme)
		return nil
	}
}
