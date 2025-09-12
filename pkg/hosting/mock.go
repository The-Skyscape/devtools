package hosting

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

// MockPlatform provides a mock implementation for testing
type MockPlatform struct {
	servers map[string]*MockServer
	mu      sync.RWMutex
	
	// Control test behavior
	shouldFail   bool
	failMessage  string
	launchCount  int
	serverCount  int
}

// NewMockPlatform creates a new mock platform for testing
func NewMockPlatform() *MockPlatform {
	return &MockPlatform{
		servers: make(map[string]*MockServer),
	}
}

// Launch creates a new mock server
func (p *MockPlatform) Launch(s *Server) (*Server, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}
	
	p.launchCount++
	p.serverCount++
	
	mockServer := &MockServer{
		id:       fmt.Sprintf("mock-server-%d", p.serverCount),
		name:     fmt.Sprintf("mock-server-%d", p.serverCount),
		ip:       fmt.Sprintf("10.0.0.%d", p.serverCount),
		env:      make(map[string]string),
		platform: p,
	}
	
	if s != nil && *s != nil {
		// Copy properties from input server if provided
		mockServer.name = (*s).GetName()
	}
	
	p.servers[mockServer.id] = mockServer
	
	var server Server = mockServer
	return &server, nil
}

// Server retrieves a mock server by ID
func (p *MockPlatform) Server(id string) (*Server, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}
	
	server, exists := p.servers[id]
	if !exists {
		return nil, fmt.Errorf("server not found: %s", id)
	}
	
	var s Server = server
	return &s, nil
}

// Test control methods

// FailNext causes the next operation to fail
func (p *MockPlatform) FailNext(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.shouldFail = true
	p.failMessage = message
}

// Reset clears the failure state
func (p *MockPlatform) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.shouldFail = false
	p.failMessage = ""
}

// GetLaunchCount returns the number of times Launch was called
func (p *MockPlatform) GetLaunchCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.launchCount
}

// GetServers returns all mock servers
func (p *MockPlatform) GetServers() []*MockServer {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	servers := make([]*MockServer, 0, len(p.servers))
	for _, s := range p.servers {
		servers = append(servers, s)
	}
	return servers
}

// MockServer provides a mock server implementation
type MockServer struct {
	id       string
	name     string
	ip       string
	env      map[string]string
	mu       sync.RWMutex
	platform *MockPlatform
	
	// Track operations
	execCalls [][]string
	copyCalls []copyCall
	aliases   []aliasCall
	destroyed bool
}

type copyCall struct {
	src, dst string
}

type aliasCall struct {
	sub, domain string
}

// GetID returns the server ID
func (s *MockServer) GetID() string {
	return s.id
}

// GetIP returns the server IP
func (s *MockServer) GetIP() string {
	return s.ip
}

// GetName returns the server name
func (s *MockServer) GetName() string {
	return s.name
}

// Launch simulates server launch
func (s *MockServer) Launch(opts ...LaunchOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.platform.shouldFail {
		return fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	// Apply launch options (if needed in future)
	for range opts {
		// Options would be applied here
	}
	
	return nil
}

// Destroy simulates server destruction
func (s *MockServer) Destroy(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.platform.shouldFail {
		return fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	s.destroyed = true
	
	// Remove from platform
	s.platform.mu.Lock()
	delete(s.platform.servers, s.id)
	s.platform.mu.Unlock()
	
	return nil
}

// Alias simulates domain aliasing
func (s *MockServer) Alias(sub, domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.platform.shouldFail {
		return fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	s.aliases = append(s.aliases, aliasCall{sub: sub, domain: domain})
	return nil
}

// Env sets an environment variable
func (s *MockServer) Env(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.platform.shouldFail {
		return fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	s.env[key] = value
	return nil
}

// Exec simulates command execution
func (s *MockServer) Exec(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var stdout, stderr bytes.Buffer
	
	if s.platform.shouldFail {
		return stdout, stderr, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	s.execCalls = append(s.execCalls, args)
	
	// Simulate some output
	stdout.WriteString(fmt.Sprintf("Mock execution of: %v\n", args))
	
	return stdout, stderr, nil
}

// Copy simulates file copying
func (s *MockServer) Copy(src, dst string) (bytes.Buffer, bytes.Buffer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var stdout, stderr bytes.Buffer
	
	if s.platform.shouldFail {
		return stdout, stderr, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	s.copyCalls = append(s.copyCalls, copyCall{src: src, dst: dst})
	
	stdout.WriteString(fmt.Sprintf("Copied %s to %s\n", src, dst))
	
	return stdout, stderr, nil
}

// Dump simulates writing data to a file
func (s *MockServer) Dump(path string, data []byte) (bytes.Buffer, bytes.Buffer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var stdout, stderr bytes.Buffer
	
	if s.platform.shouldFail {
		return stdout, stderr, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	stdout.WriteString(fmt.Sprintf("Dumped %d bytes to %s\n", len(data), path))
	
	return stdout, stderr, nil
}

// Connect simulates an interactive connection
func (s *MockServer) Connect(stdin io.Reader, stdout io.Writer, stderr io.Writer, args ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.platform.shouldFail {
		return fmt.Errorf("mock error: %s", s.platform.failMessage)
	}
	
	// Simulate connection
	fmt.Fprintf(stdout, "Connected to mock server %s\n", s.id)
	
	return nil
}

// Test helper methods

// GetExecCalls returns all exec calls made to this server
func (s *MockServer) GetExecCalls() [][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.execCalls
}

// GetEnv returns the environment variables
func (s *MockServer) GetEnv() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	env := make(map[string]string)
	for k, v := range s.env {
		env[k] = v
	}
	return env
}

// IsDestroyed returns whether the server was destroyed
func (s *MockServer) IsDestroyed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.destroyed
}