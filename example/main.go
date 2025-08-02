package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/The-Skyscape/devtools/example/controllers"
	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/authentication"
)

//go:embed all:views
var views embed.FS

func main() {
	// Create the authentication controller
	auth := models.Auth.Controller(
		authentication.WithCookie(cmp.Or(os.Getenv("TOKEN"), "example")),
		authentication.WithSignoutURL("/"),
	)

	// Create the application
	application.Serve(views,
		application.WithHostPrefix(os.Getenv("PREFIX")),
		application.WithDaisyTheme(os.Getenv("THEME")),
		application.WithController("auth", auth),
		application.WithController(controllers.Ducks()),
	)
}
