package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/The-Skyscape/devtools/example/controllers"
	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/database/local"
)

// TestDucksController demonstrates testing patterns for controllers
func TestDucksController(t *testing.T) {
	// PATTERN: Setup test database
	models.DB = local.Database(":memory:")

	// PATTERN: Create test app
	app := application.New(nil)

	// PATTERN: Setup controller
	prefix, controller := controllers.Ducks()
	controller.Setup(app)

	// Register with app
	app.Use(prefix)

	t.Run("AllDucks returns empty list initially", func(t *testing.T) {
		// PATTERN: Create controller instance with test request
		req := httptest.NewRequest("GET", "/", nil)
		ctrl := controller.Handle(req).(*controllers.DucksController)

		ducks, err := ctrl.AllDucks()
		if err != nil {
			t.Fatalf("AllDucks() error = %v", err)
		}

		if len(ducks) != 0 {
			t.Errorf("Expected 0 ducks, got %d", len(ducks))
		}
	})

	t.Run("IsEmpty returns true when no ducks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		ctrl := controller.Handle(req).(*controllers.DucksController)

		if !ctrl.IsEmpty() {
			t.Error("IsEmpty() = false, want true")
		}
	})

	t.Run("CountDucks returns zero initially", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		ctrl := controller.Handle(req).(*controllers.DucksController)

		count := ctrl.CountDucks()
		if count != 0 {
			t.Errorf("CountDucks() = %d, want 0", count)
		}
	})
}

// TestDuckModel demonstrates testing patterns for models
func TestDuckModel(t *testing.T) {
	// PATTERN: Setup test database
	models.DB = local.Database(":memory:")

	t.Run("Insert creates new duck", func(t *testing.T) {
		duck := &models.Duck{
			Name:  "TestDuck",
			Breed: "mallard",
		}

		inserted, err := models.Ducks.Insert(duck)
		if err != nil {
			t.Fatalf("Insert() error = %v", err)
		}

		if inserted.ID == "" {
			t.Error("Expected ID to be set")
		}

		if inserted.Name != "TestDuck" {
			t.Errorf("Name = %s, want TestDuck", inserted.Name)
		}
	})

	t.Run("Get retrieves duck by ID", func(t *testing.T) {
		// Create a duck first
		duck, _ := models.Ducks.Insert(&models.Duck{
			Name:  "GetTest",
			Breed: "rubber",
		})

		// Retrieve it
		retrieved, err := models.Ducks.Get(duck.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if retrieved.Name != "GetTest" {
			t.Errorf("Name = %s, want GetTest", retrieved.Name)
		}
	})

	t.Run("Search finds ducks", func(t *testing.T) {
		// Clear database
		models.DB = local.Database(":memory:")

		// Insert test ducks
		models.Ducks.Insert(&models.Duck{Name: "Donald", Breed: "mallard"})
		models.Ducks.Insert(&models.Duck{Name: "Daffy", Breed: "redbone"})

		// Search for mallards
		ducks, err := models.Ducks.Search("WHERE Breed = ?", "mallard")
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(ducks) != 1 {
			t.Errorf("Expected 1 mallard, got %d", len(ducks))
		}

		if ducks[0].Name != "Donald" {
			t.Errorf("Expected Donald, got %s", ducks[0].Name)
		}
	})
}

// TestHTMXPatterns demonstrates testing HTMX responses
func TestHTMXPatterns(t *testing.T) {
	// PATTERN: Test HTMX refresh behavior
	t.Run("Refresh sets HX-Refresh header", func(t *testing.T) {
		app := application.New(nil)
		ctrl := &controllers.DucksController{}
		ctrl.Setup(app)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", nil)
		r.Header.Set("HX-Request", "true")

		// Simulate controller refresh
		ctrl.SetRequest(r)
		ctrl.Refresh(w, r)

		// Check for HX-Refresh header
		if w.Header().Get("HX-Refresh") != "true" {
			t.Error("Expected HX-Refresh header to be set")
		}

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}
	})

	t.Run("Form submission with validation", func(t *testing.T) {
		models.DB = local.Database(":memory:")
		app := application.New(nil)

		_, controller := controllers.Ducks()
		controller.Setup(app)

		// Create form data
		form := url.Values{}
		form.Add("name", "TestDuck")
		form.Add("breed", "mallard")

		r := httptest.NewRequest("POST", "/ducks", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("HX-Request", "true")

		// Would normally route through the mux, but for testing we call directly
		// This is a limitation of the current test setup
	})
}

// TestPatternGuidance shows patterns for AI assistants
func TestPatternGuidance(t *testing.T) {
	t.Run("Controller creation pattern", func(t *testing.T) {
		// PATTERN: Controllers always return (prefix, instance)
		prefix, ctrl := controllers.Ducks()

		if prefix != "ducks" {
			t.Errorf("Expected prefix 'ducks', got %s", prefix)
		}

		if ctrl == nil {
			t.Error("Expected controller instance, got nil")
		}
	})

	t.Run("Request isolation pattern", func(t *testing.T) {
		_, controller := controllers.Ducks()

		// PATTERN: Each request gets its own controller copy
		req1 := httptest.NewRequest("GET", "/1", nil)
		req2 := httptest.NewRequest("GET", "/2", nil)

		ctrl1 := controller.Handle(req1)
		ctrl2 := controller.Handle(req2)

		// These should be different instances
		if ctrl1 == ctrl2 {
			t.Error("Expected different controller instances per request")
		}
	})
}
