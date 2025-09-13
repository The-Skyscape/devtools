package mocks

import (
	"fmt"
	"time"

	"github.com/The-Skyscape/devtools/pkg/database"
)

// MockDatabase implements the database.Database interface for testing
type MockDatabase struct {
	Models      map[string]database.Model
	QueryFunc   func(string, ...any) *database.Iter
	Error       error
	CallCount   int
	LastQuery   string
	LastArgs    []any
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Models: make(map[string]database.Model),
	}
}

// Model returns a model with default values
func (m *MockDatabase) Model() database.Model {
	m.CallCount++
	return database.Model{
		DB:        m,
		ID:        fmt.Sprintf("mock_%d", time.Now().UnixNano()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewModel creates a new model with the given ID
func (m *MockDatabase) NewModel(id string) database.Model {
	m.CallCount++
	model := database.Model{
		DB:        m,
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.Models[id] = model
	return model
}

// Query executes a query and returns an iterator
func (m *MockDatabase) Query(query string, args ...any) *database.Iter {
	m.CallCount++
	m.LastQuery = query
	m.LastArgs = args
	
	if m.QueryFunc != nil {
		return m.QueryFunc(query, args...)
	}
	
	// Return empty iterator by default
	return &database.Iter{}
}

// SetError sets an error to be returned by operations
func (m *MockDatabase) SetError(err error) {
	m.Error = err
}

// Reset clears the mock state
func (m *MockDatabase) Reset() {
	m.Models = make(map[string]database.Model)
	m.Error = nil
	m.CallCount = 0
	m.LastQuery = ""
	m.LastArgs = nil
}

// AssertCalled verifies the database was called n times
func (m *MockDatabase) AssertCalled(n int) error {
	if m.CallCount != n {
		return fmt.Errorf("expected %d calls, got %d", n, m.CallCount)
	}
	return nil
}

// AssertQuery verifies the last query matches expected
func (m *MockDatabase) AssertQuery(expected string) error {
	if m.LastQuery != expected {
		return fmt.Errorf("expected query %q, got %q", expected, m.LastQuery)
	}
	return nil
}