package mocks

import (
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// MockHandler implements the application.Handler interface for testing
type MockHandler struct {
	HandleFunc func(*http.Request) application.Handler
	SetupFunc  func(*application.App)
	Request    *http.Request
	App        *application.App
	CallCount  int
	SetupCalled bool
}

// NewMockHandler creates a new mock handler
func NewMockHandler() *MockHandler {
	return &MockHandler{}
}

// Setup implements the Handler interface
func (m *MockHandler) Setup(app *application.App) {
	m.SetupCalled = true
	m.App = app
	
	if m.SetupFunc != nil {
		m.SetupFunc(app)
	}
}

// Handle implements the Handler interface
func (m *MockHandler) Handle(req *http.Request) application.Handler {
	m.CallCount++
	m.Request = req
	
	if m.HandleFunc != nil {
		return m.HandleFunc(req)
	}
	
	// Return self by default
	return m
}

// Reset clears the mock state
func (m *MockHandler) Reset() {
	m.Request = nil
	m.App = nil
	m.CallCount = 0
	m.SetupCalled = false
}

// AssertCalled verifies the handler was called n times
func (m *MockHandler) AssertCalled(n int) bool {
	return m.CallCount == n
}

// AssertRequest verifies the request was set
func (m *MockHandler) AssertRequest() bool {
	return m.Request != nil
}