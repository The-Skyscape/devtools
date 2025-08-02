package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/authentication"
)

//go:embed all:views
var views embed.FS

func main() {
	application.Serve(views,
		application.WithHostPrefix(os.Getenv("PREFIX")),
		application.WithDaisyTheme(os.Getenv("THEME")),

		application.WithController("auth", models.Auth.Controller(
			authentication.WithCookie(cmp.Or(os.Getenv("TOKEN"), "armory")),
			authentication.WithSignoutURL("/"),
		)),

		application.WithController("code", new(controllers.CodeController)),
		application.WithController("users", new(controllers.UsersController)),

		application.WithController(admin.Panel(models.DB)),
	)
}
