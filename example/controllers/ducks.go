package controllers

import (
	"net/http"

	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// Ducks creates the ducks controller
// Factory pattern: returns (prefix, controller) for registration
func Ducks() (string, *DucksController) {
	return "ducks", &DucksController{}
}

// DucksController handles all duck-related operations
type DucksController struct {
	application.Controller // Always embed for base functionality

	// DucksCollection allows dependency injection for testing
	// If nil, defaults to models.Ducks
	DucksCollection interface {
		Get(string) (*models.Duck, error)
		Search(string, ...any) ([]*models.Duck, error)
		Count(string, ...any) int
		Insert(*models.Duck) (*models.Duck, error)
		Delete(*models.Duck) error
	}
}

// Setup registers routes (called once at startup)
func (c *DucksController) Setup(app *application.App) {
	c.Controller.Setup(app) // Always call parent first

	// Routes
	http.Handle("GET /", app.Serve("dashboard.html", nil))
	http.Handle("POST /ducks", app.ProtectFunc(c.createDuck, nil))
	http.Handle("POST /ducks/{id}/delete", app.ProtectFunc(c.deleteDuck, nil))
}

// Handle creates per-request controller copy (for thread safety)
func (c DucksController) Handle(req *http.Request) application.Handler {
	c.Request = req // Store request in the copy
	return &c       // Return pointer to copy
}

// AllDucks returns all ducks ordered by creation date
// Returns empty slice on error to prevent template execution halt
func (c *DucksController) AllDucks() []*models.Duck {
	ducks, err := models.Ducks.Search("ORDER BY CreatedAt DESC")
	if err != nil {
		// Log error but return empty slice for template safety
		// This prevents template execution from halting
		return []*models.Duck{}
	}
	return ducks
}

// CountDucks returns total duck count
func (c *DucksController) CountDucks() int {
	return models.Ducks.Count("")
}

// IsEmpty checks if there are no ducks
func (c *DucksController) IsEmpty() bool {
	return models.Ducks.Count("") == 0
}

// HTTP Handlers (private methods for routes)

// createDuck handles POST /ducks
func (c *DucksController) createDuck(w http.ResponseWriter, r *http.Request) {
	// Validate input
	validator := c.Validator()
	validator.CheckRequired("name", r.FormValue("name"))
	validator.CheckLength("name", r.FormValue("name"), 2, 50)
	validator.CheckRequired("color", r.FormValue("color"))

	if err := validator.Result(); err != nil {
		c.RenderError(w, r, err)
		return
	}

	// Create duck
	duck := &models.Duck{
		Name:        r.FormValue("name"),
		Color:       r.FormValue("color"),
		Description: r.FormValue("description"),
		UserID:      "", // Would be from auth.CurrentUser()
	}

	if _, err := models.Ducks.Insert(duck); err != nil {
		c.RenderError(w, r, err)
		return
	}

	// Send email notification (demonstrates email API)
	go func() {
		models.Emails.Send(
			"user@example.com", // Would be user's email
			"New Duck Created!",
			emailing.WithTemplate("duck-created.html"),
			emailing.WithRequest(r),
			emailing.WithData("DuckName", duck.Name),
			emailing.WithData("DuckColor", duck.Color),
			emailing.WithData("DuckDescription", duck.Description),
		)
	}()

	c.Refresh(w, r) // HTMX full page refresh
}

// deleteDuck handles POST /ducks/{id}/delete
func (c *DucksController) deleteDuck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	duck, err := models.Ducks.Get(id)
	if err != nil {
		c.RenderError(w, r, application.ErrNotFound)
		return
	}

	if err := models.Ducks.Delete(duck); err != nil {
		c.RenderError(w, r, err)
		return
	}

	c.Redirect(w, r, "/")
}
