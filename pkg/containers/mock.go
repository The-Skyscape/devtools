package containers

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

// MockHost provides a mock implementation of Host for testing
type MockHost struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	
	// Test control
	execCalls    [][]string
	shouldFail   bool
	failMessage  string
	execOutput   map[string]string // Map command to output
	mu           sync.RWMutex
}

// NewMockHost creates a new mock host for testing
func NewMockHost() *MockHost {
	return &MockHost{
		execCalls:  [][]string{},
		execOutput: make(map[string]string),
		stdout:     &bytes.Buffer{},
		stderr:     &bytes.Buffer{},
	}
}

// SetStdin sets the standard input
func (h *MockHost) SetStdin(r io.Reader) {
	h.stdin = r
}

// SetStdout sets the standard output
func (h *MockHost) SetStdout(w io.Writer) {
	h.stdout = w
}

// SetStderr sets the standard error
func (h *MockHost) SetStderr(w io.Writer) {
	h.stderr = w
}

// Exec simulates command execution
func (h *MockHost) Exec(args ...string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.execCalls = append(h.execCalls, args)
	
	if h.shouldFail {
		return fmt.Errorf("mock error: %s", h.failMessage)
	}
	
	// Check for predefined output
	cmdKey := strings.Join(args, " ")
	if output, ok := h.execOutput[cmdKey]; ok {
		if h.stdout != nil {
			h.stdout.Write([]byte(output))
		}
	} else {
		// Default output
		if h.stdout != nil {
			fmt.Fprintf(h.stdout, "Mock execution: %s\n", cmdKey)
		}
	}
	
	return nil
}

// Dump simulates writing a file (no-op for mock)
func (h *MockHost) Dump(path string, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.shouldFail {
		return fmt.Errorf("mock error: %s", h.failMessage)
	}

	// Record the dump as an exec call for testing purposes
	h.execCalls = append(h.execCalls, []string{"dump", path, string(data)})
	return nil
}

// Test control methods

// FailNext causes the next exec to fail
func (h *MockHost) FailNext(message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.shouldFail = true
	h.failMessage = message
}

// Reset clears the failure state
func (h *MockHost) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.shouldFail = false
	h.failMessage = ""
	h.execCalls = [][]string{}
}

// SetExecOutput sets the output for a specific command
func (h *MockHost) SetExecOutput(command string, output string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.execOutput[command] = output
}

// GetExecCalls returns all exec calls
func (h *MockHost) GetExecCalls() [][]string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	calls := make([][]string, len(h.execCalls))
	copy(calls, h.execCalls)
	return calls
}

// MockService provides a mock implementation of Service for testing
type MockService struct {
	*Service
	
	// Track operations
	startCalls   int
	stopCalls    int
	restartCalls int
	isRunning    bool
	logs         []string
	mu           sync.RWMutex
}

// NewMockService creates a new mock service for testing
func NewMockService(name string) *MockService {
	return &MockService{
		Service: &Service{
			Host:         NewMockHost(),
			Name:         name,
			Image:        "mock-image:latest",
			Status:       "created",
			Ports:        make(map[int]int),
			Mounts:       make(map[string]string),
			Env:          make(map[string]string),
		},
		isRunning: false,
		logs:      []string{},
	}
}

// Start simulates starting the service
func (s *MockService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if host, ok := s.Host.(*MockHost); ok && host.shouldFail {
		return fmt.Errorf("mock error: %s", host.failMessage)
	}
	
	s.startCalls++
	s.isRunning = true
	s.Status = "running"
	s.logs = append(s.logs, fmt.Sprintf("Service %s started", s.Name))
	
	return nil
}

// Stop simulates stopping the service
func (s *MockService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if host, ok := s.Host.(*MockHost); ok && host.shouldFail {
		return fmt.Errorf("mock error: %s", host.failMessage)
	}
	
	s.stopCalls++
	s.isRunning = false
	s.Status = "stopped"
	s.logs = append(s.logs, fmt.Sprintf("Service %s stopped", s.Name))
	
	return nil
}

// Restart simulates restarting the service
func (s *MockService) Restart() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if host, ok := s.Host.(*MockHost); ok && host.shouldFail {
		return fmt.Errorf("mock error: %s", host.failMessage)
	}
	
	s.restartCalls++
	s.isRunning = true
	s.Status = "running"
	s.logs = append(s.logs, fmt.Sprintf("Service %s restarted", s.Name))
	
	return nil
}

// IsRunning returns whether the service is running
func (s *MockService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetLogs returns the service logs
func (s *MockService) GetLogs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	logs := make([]string, len(s.logs))
	copy(logs, s.logs)
	return logs
}

// GetStartCount returns how many times Start was called
func (s *MockService) GetStartCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startCalls
}

// GetStopCount returns how many times Stop was called
func (s *MockService) GetStopCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stopCalls
}

// AddLog adds a log entry (for testing)
func (s *MockService) AddLog(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, message)
}