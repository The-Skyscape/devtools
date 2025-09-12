package testutils

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/database"
	_ "github.com/mattn/go-sqlite3"
)

// TestDatabase creates an in-memory SQLite database for testing
type TestDatabase struct {
	*sql.DB
	t *testing.T
}

// NewTestDatabase creates a new test database
func NewTestDatabase(t *testing.T) *TestDatabase {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	return &TestDatabase{
		DB: db,
		t:  t,
	}
}

// Close cleans up the test database
func (db *TestDatabase) Close() {
	if db.DB != nil {
		db.DB.Close()
	}
}

// CreateTable creates a table for a model (Entity interface)
func (db *TestDatabase) CreateTable(model database.Entity) error {
	// Get table name
	tableName := model.Table()

	// Build CREATE TABLE statement based on model fields
	// This is a simplified version - production would use reflection
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			ID TEXT PRIMARY KEY,
			CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, tableName)

	_, err := db.Exec(query)
	return err
}

// Repository creates a repository for a model type
// Note: This requires a proper DynamicDB instance, not just sql.DB
// For testing models, we'll need to mock this properly
func (db *TestDatabase) Repository(model database.Entity) any {
	// This would need a proper implementation for actual database testing
	// For now, return nil as we're testing business logic, not database operations
	return nil
}

// Seed adds test data to the database
func (db *TestDatabase) Seed(table string, data map[string]any) (string, error) {
	// Build INSERT statement
	columns := make([]string, 0, len(data))
	values := make([]any, 0, len(data))
	placeholders := make([]string, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		joinStrings(columns, ", "),
		joinStrings(placeholders, ", "),
	)

	result, err := db.Exec(query, values...)
	if err != nil {
		return "", err
	}

	// Get the last inserted ID (for SQLite)
	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", id), nil
}

// ExecuteSQL runs arbitrary SQL for test setup
func (db *TestDatabase) ExecuteSQL(query string, args ...any) error {
	_, err := db.Exec(query, args...)
	return err
}

// AssertCount verifies the number of rows in a table
func (db *TestDatabase) AssertCount(table string, expected int, where string, args ...any) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if where != "" {
		query += " WHERE " + where
	}

	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		db.t.Fatalf("Failed to count rows: %v", err)
	}

	if count != expected {
		db.t.Errorf("Expected %d rows in %s, got %d", expected, table, count)
	}
}

// helper function to join strings
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
