package controllers

import (
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/authentication"
)

// Auth returns the auth controller with embedded authentication
func Auth() (string, *AuthController) {
	return "auth", &AuthController{
		Controller: authentication.New(),
	}
}

// AuthController embeds authentication.Controller
type AuthController struct {
	application.Controller
	*authentication.Controller // Embedded controller with CurrentUser, IsAuthenticated, etc.
}

// Setup initializes the controller
func (c *AuthController) Setup(app *application.App) {
	c.Controller.Setup(app)
	c.Controller.Setup(app)
}

// Handle returns the controller instance
func (c *AuthController) Handle(r *http.Request) application.Handler {
	c.Request = r
	c.Controller.Handle(r)
	return c
}

// CustomAuthMethod is a custom method on this controller
func (c *AuthController) CustomAuthMethod() bool {
	return true
}
