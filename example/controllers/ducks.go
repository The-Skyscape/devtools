package controllers

import (
	"net/http"

	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/application"
)

// Ducks is a factory function with the prefix and instance
func Ducks() (string, *DucksController) {
	return "ducks", &DucksController{}
}

// DucksController is the controller for the ducks
type DucksController struct {
	application.BaseController
}

// Setup is called when the application is started
func (c *DucksController) Setup(app *application.App) {
	c.BaseController.Setup(app)

	http.Handle("GET /", app.Serve("dashboard.html", nil))
	http.Handle("POST /", app.ProtectFunc(c.spawnDuck, nil))
}

// Handle is called when each request is handled
func (c DucksController) Handle(req *http.Request) application.Controller {
	c.Request = req
	return &c
}

// AllDucks is a function that can be called in views
func (c *DucksController) AllDucks() ([]*models.Duck, error) {
	return models.Ducks.Search("")
}

// spawnDuck is a HandlerFunc that is called when the user submits a duck
func (c *DucksController) spawnDuck(w http.ResponseWriter, r *http.Request) {
	// Creating a instance model of a duck
	duck := &models.Duck{
		Name:  r.FormValue("name"),
		Breed: r.FormValue("breed"),
	}

	// Saving ducks to the ducks collection
	if _, err := models.Ducks.Insert(duck); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	// Refreshing the page via HTMX integration
	c.Refresh(w, r)
}
