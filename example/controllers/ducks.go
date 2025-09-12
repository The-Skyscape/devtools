package controllers

import (
	"net/http"

	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/application"
)

// Ducks returns the controller prefix and instance
// PATTERN: Factory function ALWAYS returns (string, *Controller)
// AI: The string becomes the template function name (e.g., "ducks" → {{ducks.Method}})
func Ducks() (string, *DucksController) {
	return "ducks", &DucksController{}
}

// DucksController manages duck-related operations
// PATTERN: Always embed application.Controller (never BaseController)
type DucksController struct {
	application.Controller
}

// Setup registers routes and initializes the controller
// PATTERN: Called ONCE at application startup, not per request
// AI: Register all routes here, don't create routes elsewhere
func (c *DucksController) Setup(app *application.App) {
	c.Controller.Setup(app) // PATTERN: Always call parent Setup first

	// PATTERN: GET routes typically render pages
	http.Handle("GET /", app.Serve("dashboard.html", nil))
	http.Handle("GET /ducks/{id}", app.Serve("duck-detail.html", nil))

	// PATTERN: POST routes typically modify data then refresh/redirect
	http.Handle("POST /ducks", app.ProtectFunc(c.spawnDuck, nil))
	http.Handle("POST /ducks/{id}/update", app.ProtectFunc(c.updateDuck, nil))
	http.Handle("POST /ducks/{id}/delete", app.ProtectFunc(c.deleteDuck, nil))
}

// Handle creates a request-scoped controller instance
// PATTERN: MUST use VALUE receiver (not pointer) for request isolation
// AI: This creates a COPY for each request - no shared state!
func (c DucksController) Handle(req *http.Request) application.Handler {
	c.Request = req // Modifies the copy, not the original
	return &c       // Returns pointer to the copy
}

// AllDucks returns all ducks for template display
// PATTERN: Public methods (Capitalized) are accessible in templates
// AI: This is called in templates as {{ducks.AllDucks}}
func (c *DucksController) AllDucks() ([]*models.Duck, error) {
	// PATTERN: Use PascalCase in SQL (CreatedAt not created_at)
	return models.Ducks.Search("ORDER BY CreatedAt DESC")
}

// GetDuck returns a single duck by ID from the URL
// PATTERN: Template methods - use c.PathValue() since request is from Handle()
func (c *DucksController) GetDuck() (*models.Duck, error) {
	// PATTERN: In template methods, use c.PathValue() (request from Handle)
	// PATTERN: IDs are ALWAYS strings
	id := c.PathValue("id")
	if id == "" {
		return nil, application.ErrNotFound
	}
	return models.Ducks.Get(id)
}

// CountDucks returns the total number of ducks
// PATTERN: Provide computed values through methods, not in templates
func (c *DucksController) CountDucks() int {
	return models.Ducks.Count("")
}

// spawnDuck handles POST /ducks - creates a new duck
// PATTERN: Private methods (lowercase) are HTTP handlers
// AI: These are NOT accessible in templates, only through routes
func (c *DucksController) spawnDuck(w http.ResponseWriter, r *http.Request) {
	c.SetRequest(r) // PATTERN: Always set request first in handlers

	// PATTERN: Use Validator for input validation
	validator := c.Validator()
	validator.CheckRequired("name", r.FormValue("name"))
	validator.CheckLength("name", r.FormValue("name"), 2, 50)

	// Validate breed is one of allowed values
	breed := r.FormValue("breed")
	if breed != "mallard" && breed != "redbone" && breed != "rubber" {
		validator.AddError("breed", "Invalid breed selected")
	}

	// PATTERN: Always check validation result
	if err := validator.Result(); err != nil {
		c.RenderError(w, r, err) // PATTERN: Always render errors to user
		return
	}

	// Create the duck model
	duck := &models.Duck{
		Name:  r.FormValue("name"),
		Breed: r.FormValue("breed"),
	}

	// PATTERN: Handle database errors properly
	if _, err := models.Ducks.Insert(duck); err != nil {
		c.RenderError(w, r, err)
		return
	}

	// PATTERN: Use Refresh for HTMX page updates
	// This triggers an HX-Refresh header for full page reload
	c.Refresh(w, r)
}

// updateDuck handles POST /ducks/{id}/update
// PATTERN: Update existing resources
func (c *DucksController) updateDuck(w http.ResponseWriter, r *http.Request) {
	c.SetRequest(r)

	// Get the duck to update
	// PATTERN: In handlers, use r.PathValue() for clarity
	// PATTERN: IDs are ALWAYS strings
	id := r.PathValue("id")
	duck, err := models.Ducks.Get(id)
	if err != nil {
		c.RenderError(w, r, application.ErrNotFound)
		return
	}

	// Validate input
	validator := c.Validator()
	validator.CheckRequired("name", r.FormValue("name"))

	if err := validator.Result(); err != nil {
		c.RenderError(w, r, err)
		return
	}

	// Update the duck
	duck.Name = r.FormValue("name")
	duck.Breed = r.FormValue("breed")

	if err := models.Ducks.Update(duck); err != nil {
		c.RenderError(w, r, err)
		return
	}

	// PATTERN: Use Redirect for navigation after updates
	c.Redirect(w, r, "/ducks/"+string(duck.ID))
}

// deleteDuck handles POST /ducks/{id}/delete
// PATTERN: Delete resources and redirect
func (c *DucksController) deleteDuck(w http.ResponseWriter, r *http.Request) {
	c.SetRequest(r)

	// Get the duck to delete
	// PATTERN: In handlers, use r.PathValue() for clarity
	// PATTERN: IDs are ALWAYS strings
	id := r.PathValue("id")
	duck, err := models.Ducks.Get(id)
	if err != nil {
		c.RenderError(w, r, application.ErrNotFound)
		return
	}

	// Delete the duck
	if err := models.Ducks.Delete(duck); err != nil {
		c.RenderError(w, r, err)
		return
	}

	// PATTERN: Redirect after deletion
	c.Redirect(w, r, "/")
}

// IsEmpty checks if there are no ducks
// PATTERN: Provide helper methods for template logic
func (c *DucksController) IsEmpty() bool {
	return models.Ducks.Count("") == 0
}
