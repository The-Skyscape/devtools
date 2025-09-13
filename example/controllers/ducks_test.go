package controllers

import (
	"testing"

	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/engines/sqlite3"
)

func TestDucksController(t *testing.T) {
	// Setup test database
	testDB := sqlite3.Open(":memory:", nil).Dynamic()
	models.DB = testDB
	models.Ducks = database.Manage(models.DB, new(models.Duck))
	
	_, controller := Ducks()
	
	t.Run("AllDucks returns ducks in correct order", func(t *testing.T) {
		// Create ducks with different timestamps
		duck1 := &models.Duck{Name: "First", Color: "Yellow"}
		duck2 := &models.Duck{Name: "Second", Color: "Blue"}
		
		models.Ducks.Insert(duck1)
		models.Ducks.Insert(duck2)
		
		// Should return in reverse chronological order (newest first)
		ducks, err := controller.AllDucks()
		if err != nil {
			t.Fatalf("AllDucks failed: %v", err)
		}
		
		if len(ducks) != 2 {
			t.Errorf("Expected 2 ducks, got %d", len(ducks))
		}
		
		// Newest should be first due to ORDER BY CreatedAt DESC
		if ducks[0].Name != "Second" {
			t.Error("Ducks should be ordered by creation date descending")
		}
	})
	
	t.Run("CountDucks returns accurate count", func(t *testing.T) {
		// Clear database
		models.DB = sqlite3.Open(":memory:", nil).Dynamic()
		models.Ducks = database.Manage(models.DB, new(models.Duck))
		
		// Initially empty
		if count := controller.CountDucks(); count != 0 {
			t.Errorf("Expected 0 ducks initially, got %d", count)
		}
		
		// Add ducks and verify count
		models.Ducks.Insert(&models.Duck{Name: "Duck1", Color: "Red"})
		models.Ducks.Insert(&models.Duck{Name: "Duck2", Color: "Green"})
		
		if count := controller.CountDucks(); count != 2 {
			t.Errorf("Expected 2 ducks after inserts, got %d", count)
		}
	})
	
	t.Run("IsEmpty correctly identifies empty state", func(t *testing.T) {
		// Clear database
		models.DB = sqlite3.Open(":memory:", nil).Dynamic()
		models.Ducks = database.Manage(models.DB, new(models.Duck))
		
		// Should be empty initially
		if !controller.IsEmpty() {
			t.Error("IsEmpty should return true for empty database")
		}
		
		// Add a duck
		models.Ducks.Insert(&models.Duck{Name: "Lonely", Color: "Black"})
		
		// Should not be empty
		if controller.IsEmpty() {
			t.Error("IsEmpty should return false when ducks exist")
		}
	})
	
	// Note: The HTTP handlers (createDuck, deleteDuck) would require:
	// - Full HTTP request/response testing
	// - Template rendering
	// - Validation testing
	// These are integration concerns better tested with the full app context
}