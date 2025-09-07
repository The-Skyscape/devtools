package controllers

import (
	"github.com/The-Skyscape/devtools/pkg/application"
	"net/http"
)

// Home returns the home controller
func Home() (string, *HomeController) {
	return "home", &HomeController{}
}

// HomeController handles home page
type HomeController struct {
	application.BaseController
}

// Setup initializes the controller
func (c *HomeController) Setup(app *application.App) {
	c.BaseController.Setup(app)
}

// Handle returns the controller instance
func (c *HomeController) Handle(r *http.Request) application.Controller {
	c.Request = r
	return c
}

// Welcome returns a welcome message (exported method)
func (c *HomeController) Welcome() string {
	return "Welcome!"
}

// GetUserCount returns the number of users
func (c *HomeController) GetUserCount() int {
	return 42
}

// privateMethod is not exported
func (c *HomeController) privateMethod() string {
	return "private"
}
