package database_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/engines/sqlite3"
)

// Test model for repository tests
type TestUser struct {
	database.Model
	Name  string
	Email string
	Age   int
	Active bool
}

func (*TestUser) Table() string { return "test_users" }

// TestRepositoryCRUD tests all CRUD operations
func TestRepositoryCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	// Register collection
	Users := database.Manage(db, new(TestUser))
	
	// Test Insert
	user := &TestUser{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
		Active: true,
	}
	
	inserted, err := Users.Insert(user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	
	if inserted.ID == "" {
		t.Error("Expected ID to be generated")
	}
	
	if inserted.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	
	// Test Get
	retrieved, err := Users.Get(inserted.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	
	if retrieved.Name != "Alice" {
		t.Errorf("Expected name Alice, got %s", retrieved.Name)
	}
	
	// Test Update
	retrieved.Name = "Alice Smith"
	err = Users.Update(retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	
	// Verify update
	updated, err := Users.Get(retrieved.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	
	if updated.Name != "Alice Smith" {
		t.Errorf("Expected updated name, got %s", updated.Name)
	}
	
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt after update")
	}
	
	// Test Delete
	err = Users.Delete(updated)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	
	// Verify deletion
	_, err = Users.Get(updated.ID)
	if !errors.Is(err, database.ErrNotFound) {
		t.Error("Expected ErrNotFound after deletion")
	}
}

// TestRepositoryPascalCaseSQL verifies PascalCase field names in SQL
func TestRepositoryPascalCaseSQL(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	Users := database.Manage(db, new(TestUser))
	
	// Insert test data
	users := []*TestUser{
		{Name: "Bob", Email: "bob@example.com", Age: 25, Active: true},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35, Active: false},
		{Name: "Diana", Email: "diana@example.com", Age: 28, Active: true},
	}
	
	for _, u := range users {
		_, err := Users.Insert(u)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}
	
	// Test PascalCase in WHERE clause
	results, err := Users.Search("WHERE Active = ? ORDER BY Age", true)
	if err != nil {
		t.Fatalf("Search with PascalCase failed: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(results))
	}
	
	// Verify ordering by PascalCase field
	if results[0].Age > results[1].Age {
		t.Error("Results not properly ordered by Age")
	}
	
	// Test compound conditions with PascalCase
	results, err = Users.Search("WHERE Age > ? AND Active = ?", 26, true)
	if err != nil {
		t.Fatalf("Compound search failed: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	
	if results[0].Name != "Diana" {
		t.Errorf("Expected Diana, got %s", results[0].Name)
	}
}

// TestRepositoryCount tests the Count method
func TestRepositoryCount(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	Users := database.Manage(db, new(TestUser))
	
	// Initially empty
	count := Users.Count("")
	if count != 0 {
		t.Errorf("Expected 0 users initially, got %d", count)
	}
	
	// Add users
	for i := 0; i < 5; i++ {
		_, err := Users.Insert(&TestUser{
			Name:   "User",
			Email:  "user@example.com",
			Active: i%2 == 0,
		})
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}
	
	// Count all
	count = Users.Count("")
	if count != 5 {
		t.Errorf("Expected 5 users, got %d", count)
	}
	
	// Count with condition
	count = Users.Count("WHERE Active = ?", true)
	if count != 3 {
		t.Errorf("Expected 3 active users, got %d", count)
	}
}

// TestRepositoryFirst tests the First method
func TestRepositoryFirst(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	Users := database.Manage(db, new(TestUser))
	
	// Test empty result
	_, err := Users.First("WHERE Name = ?", "NonExistent")
	if !errors.Is(err, database.ErrNotFound) {
		t.Error("Expected ErrNotFound for empty result")
	}
	
	// Add users
	_, err = Users.Insert(&TestUser{Name: "First", Age: 20})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	_, err = Users.Insert(&TestUser{Name: "Second", Age: 30})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	
	// Test First with ordering
	user, err := Users.First("ORDER BY Age")
	if err != nil {
		t.Fatalf("First failed: %v", err)
	}
	
	if user.Name != "First" {
		t.Errorf("Expected First user, got %s", user.Name)
	}
	
	// Test First with condition
	user, err = Users.First("WHERE Age > ? ORDER BY Age", 25)
	if err != nil {
		t.Fatalf("First with condition failed: %v", err)
	}
	
	if user.Name != "Second" {
		t.Errorf("Expected Second user, got %s", user.Name)
	}
}

// TestRepositoryConcurrency tests concurrent operations
func TestRepositoryConcurrency(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	Users := database.Manage(db, new(TestUser))
	
	// Run concurrent inserts
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			user := &TestUser{
				Name:  "Concurrent",
				Email: "concurrent@example.com",
				Age:   index,
			}
			
			_, err := Users.Insert(user)
			if err != nil {
				errors <- err
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent insert error: %v", err)
	}
	
	// Verify all inserts succeeded
	count := Users.Count("")
	if count != 100 {
		t.Errorf("Expected 100 users after concurrent inserts, got %d", count)
	}
}

// TestRepositoryPagination tests paginated queries
func TestRepositoryPagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	Users := database.Manage(db, new(TestUser))
	
	// Insert 25 users
	for i := 1; i <= 25; i++ {
		_, err := Users.Insert(&TestUser{
			Name: "User",
			Age:  i,
		})
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}
	
	// Test first page
	page1, total, err := Users.SearchPaginated("ORDER BY Age", 10, 0)
	if err != nil {
		t.Fatalf("First page failed: %v", err)
	}
	
	if total != 25 {
		t.Errorf("Expected total 25, got %d", total)
	}
	
	if len(page1) != 10 {
		t.Errorf("Expected 10 items in first page, got %d", len(page1))
	}
	
	if page1[0].Age != 1 {
		t.Errorf("Expected first item age 1, got %d", page1[0].Age)
	}
	
	// Test second page
	page2, _, err := Users.SearchPaginated("ORDER BY Age", 10, 10)
	if err != nil {
		t.Fatalf("Second page failed: %v", err)
	}
	
	if len(page2) != 10 {
		t.Errorf("Expected 10 items in second page, got %d", len(page2))
	}
	
	if page2[0].Age != 11 {
		t.Errorf("Expected first item of page 2 age 11, got %d", page2[0].Age)
	}
	
	// Test last page
	page3, _, err := Users.SearchPaginated("ORDER BY Age", 10, 20)
	if err != nil {
		t.Fatalf("Third page failed: %v", err)
	}
	
	if len(page3) != 5 {
		t.Errorf("Expected 5 items in last page, got %d", len(page3))
	}
}

// BenchmarkRepositoryInsert benchmarks insert operations
func BenchmarkRepositoryInsert(b *testing.B) {
	db := setupTestDB(b)
	defer cleanupTestDB(b, db)
	
	Users := database.Manage(db, new(TestUser))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &TestUser{
			Name:  "Benchmark",
			Email: "bench@example.com",
			Age:   25,
		}
		_, _ = Users.Insert(user)
	}
}

// BenchmarkRepositorySearch benchmarks search operations
func BenchmarkRepositorySearch(b *testing.B) {
	db := setupTestDB(b)
	defer cleanupTestDB(b, db)
	
	Users := database.Manage(db, new(TestUser))
	
	// Insert test data
	for i := 0; i < 100; i++ {
		_, _ = Users.Insert(&TestUser{
			Name:   "User",
			Email:  "user@example.com",
			Age:    i,
			Active: i%2 == 0,
		})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Users.Search("WHERE Active = ? ORDER BY Age LIMIT 10", true)
	}
}

// BenchmarkRepositoryGet benchmarks get by ID operations
func BenchmarkRepositoryGet(b *testing.B) {
	db := setupTestDB(b)
	defer cleanupTestDB(b, db)
	
	Users := database.Manage(db, new(TestUser))
	
	// Insert a user
	user, _ := Users.Insert(&TestUser{
		Name:  "Benchmark",
		Email: "bench@example.com",
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Users.Get(user.ID)
	}
}

// Helper functions

func setupTestDB(t testing.TB) *database.DynamicDB {
	t.Helper()
	
	// Create in-memory database (pass nil for no migrations)
	sqliteDB := sqlite3.Open(":memory:", nil)
	
	// Create test table - SQLite3 embeds *sql.DB
	_, err := sqliteDB.Exec(`
		CREATE TABLE test_users (
			ID TEXT PRIMARY KEY,
			CreatedAt DATETIME,
			UpdatedAt DATETIME,
			Name TEXT,
			Email TEXT,
			Age INTEGER,
			Active BOOLEAN
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	
	// Use the Dynamic() method to get DynamicDB
	return sqliteDB.Dynamic()
}

func cleanupTestDB(t testing.TB, db *database.DynamicDB) {
	t.Helper()
	// For in-memory database, no cleanup needed
	// Connection will be closed when test ends
}