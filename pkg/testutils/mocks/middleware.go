package mocks

import (
	"net/http"
)

// MockMiddleware implements the application.Middleware interface for testing
type MockMiddleware struct {
	HandleFunc    func(http.Handler) http.Handler
	CallCount     int
	LastHandler   http.Handler
	ShouldProcess bool
	Error         error
}

// NewMockMiddleware creates a new mock middleware
func NewMockMiddleware() *MockMiddleware {
	return &MockMiddleware{
		ShouldProcess: true,
	}
}

// Handle implements the Middleware interface
func (m *MockMiddleware) Handle(next http.Handler) http.Handler {
	m.CallCount++
	m.LastHandler = next
	
	if m.HandleFunc != nil {
		return m.HandleFunc(next)
	}
	
	// Default implementation that can be controlled
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Error != nil {
			http.Error(w, m.Error.Error(), http.StatusInternalServerError)
			return
		}
		
		if m.ShouldProcess {
			// Call the next handler
			next.ServeHTTP(w, r)
		} else {
			// Block the request
			http.Error(w, "Blocked by middleware", http.StatusForbidden)
		}
	})
}

// Reset clears the mock state
func (m *MockMiddleware) Reset() {
	m.CallCount = 0
	m.LastHandler = nil
	m.ShouldProcess = true
	m.Error = nil
}

// SetError sets an error to be returned
func (m *MockMiddleware) SetError(err error) {
	m.Error = err
}

// Block sets the middleware to block requests
func (m *MockMiddleware) Block() {
	m.ShouldProcess = false
}

// Allow sets the middleware to allow requests
func (m *MockMiddleware) Allow() {
	m.ShouldProcess = true
}

// AssertCalled verifies the middleware was called n times
func (m *MockMiddleware) AssertCalled(n int) bool {
	return m.CallCount == n
}