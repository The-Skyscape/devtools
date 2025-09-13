package tests

import (
	"sync"

	"github.com/The-Skyscape/devtools/example/models"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/engines/sqlite3"
)

var (
	// DB is the test database instance
	DB *database.DynamicDB
	
	// TestDucks is the test collection for ducks
	TestDucks *database.Collection[*models.Duck]
	
	// once ensures models are registered only once
	once sync.Once
)

// init sets up the test database
func init() {
	// Create in-memory test database
	testDB := sqlite3.Open(":memory:", nil)
	DB = testDB.Dynamic()
	
	// Register models once
	once.Do(func() {
		registerTestModels()
	})
}

// RegisterTestModel registers a model with the test database
func RegisterTestModel[E database.Entity](model E) *database.Collection[E] {
	return database.Manage(DB, model)
}

// registerTestModels registers all models with the test database
func registerTestModels() {
	// Register Duck model
	TestDucks = RegisterTestModel(new(models.Duck))
}

// InsertTestDucks creates standard test data
func InsertTestDucks() {
	// Clear any existing data
	DB.Query("DELETE FROM ducks").Exec()
	
	// Create test ducks
	TestDucks.Insert(&models.Duck{Name: "Donald", Color: "white"})
	TestDucks.Insert(&models.Duck{Name: "Daffy", Color: "black"})
	TestDucks.Insert(&models.Duck{Name: "Howard", Color: "brown"})
}

// ClearTestData removes all test data
func ClearTestData() {
	DB.Query("DELETE FROM ducks").Exec()
}