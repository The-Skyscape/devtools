package main

import (
	"cmp"
	"embed"
	"log"
	"os"

	"github.com/The-Skyscape/devtools/example/controllers"
	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/authentication"
)

var (
	//go:embed all:views
	views embed.FS

	//go:embed all:emails
	emails embed.FS
)

func main() {
	// Load email templates from embedded filesystem
	if err := models.Emails.LoadTemplates(emails); err != nil {
		log.Fatal("Failed to load email templates:", err)
	}

	// Setup authentication
	auth := models.Auth.Controller(
		authentication.WithCookie(cmp.Or(os.Getenv("TOKEN"), "example")),
		authentication.WithSignoutURL("/"),
	)

	// Start the application
	application.Serve(views,
		application.WithHostPrefix(os.Getenv("PREFIX")),
		application.WithDaisyTheme(cmp.Or(os.Getenv("THEME"), "corporate")),
		application.WithController("auth", auth),
		application.WithController(controllers.Ducks()),
	)
}
