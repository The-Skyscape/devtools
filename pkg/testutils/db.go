package testutils

import (
	"testing"
	
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/engines/sqlite3"
)

// TestDB creates an in-memory test database
func TestDB() *database.DynamicDB {
	db := sqlite3.Open(":memory:", nil)
	return db.Dynamic()
}

// WithTestDB runs a test with a fresh in-memory database
func WithTestDB(t *testing.T, f func(*database.DynamicDB)) {
	db := TestDB()
	f(db)
}