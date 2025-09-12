package testutils

import (
	"time"
)

// MockTime provides time manipulation for testing
type MockTime struct {
	current time.Time
	frozen  bool
}

// NewMockTime creates a new mock time controller
func NewMockTime() *MockTime {
	return &MockTime{
		current: time.Now(),
		frozen:  false,
	}
}

// Now returns the current mock time
func (m *MockTime) Now() time.Time {
	if m.frozen {
		return m.current
	}
	return time.Now()
}

// Set sets the mock time to a specific value
func (m *MockTime) Set(t time.Time) {
	m.current = t
	m.frozen = true
}

// Advance moves the mock time forward by a duration
func (m *MockTime) Advance(d time.Duration) {
	if !m.frozen {
		m.current = time.Now()
		m.frozen = true
	}
	m.current = m.current.Add(d)
}

// Reset unfreezes time
func (m *MockTime) Reset() {
	m.frozen = false
}

// TimeTravel temporarily sets time for a function
func TimeTravel(target time.Time, f func()) {
	// This is a simplified version - in production you'd need
	// to hook into the actual time package
	f()
}

// WithTimeout runs a function with a timeout
func WithTimeout(t interface{ Fatal(...any) }, timeout time.Duration, f func()) {
	done := make(chan bool, 1)
	go func() {
		f()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(timeout):
		t.Fatal("Test timed out after", timeout)
	}
}
