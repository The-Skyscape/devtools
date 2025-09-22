package mock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/The-Skyscape/devtools/pkg/hosting"
)

// Platform provides a mock hosting platform for testing
type Platform struct {
	servers map[string]*Server
	volumes map[string]*Volume
	domains map[string]*Domain
	mu      sync.RWMutex

	// Test control
	shouldFail  bool
	failMessage string

	// Counters
	serverCount int
	volumeCount int
	domainCount int
}

// New creates a new mock platform
func New() *Platform {
	return &Platform{
		servers: make(map[string]*Server),
		volumes: make(map[string]*Volume),
		domains: make(map[string]*Domain),
	}
}

// Launch creates a new mock server
func (p *Platform) Launch(opts ...interface{}) (hosting.Server, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	p.serverCount++
	server := &Server{
		id:       fmt.Sprintf("srv_%d", p.serverCount),
		name:     fmt.Sprintf("server-%d", p.serverCount),
		ip:       fmt.Sprintf("10.0.0.%d", p.serverCount),
		platform: p,
		volumes:  make(map[string]*Volume),
		domains:  make(map[string]*Domain),
	}

	// Apply options directly to the server
	for _, opt := range opts {
		if fn, ok := opt.(ServerOption); ok {
			fn(server)
		}
	}

	p.servers[server.id] = server

	return server, nil
}

// Server retrieves a server by ID
func (p *Platform) Server(id string) (hosting.Server, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	server, exists := p.servers[id]
	if !exists {
		return nil, fmt.Errorf("server not found: %s", id)
	}

	return server, nil
}

// Servers returns all servers
func (p *Platform) Servers() ([]hosting.Server, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	servers := make([]hosting.Server, 0, len(p.servers))
	for _, s := range p.servers {
		servers = append(servers, s)
	}
	return servers, nil
}

// Volume retrieves a volume by ID
func (p *Platform) Volume(id string) (hosting.Volume, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	volume, exists := p.volumes[id]
	if !exists {
		return nil, fmt.Errorf("volume not found: %s", id)
	}

	return volume, nil
}

// Volumes returns all volumes
func (p *Platform) Volumes() ([]hosting.Volume, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	volumes := make([]hosting.Volume, 0, len(p.volumes))
	for _, v := range p.volumes {
		volumes = append(volumes, v)
	}
	return volumes, nil
}

// Domain retrieves a domain by name
func (p *Platform) Domain(name string) (hosting.Domain, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	domain, exists := p.domains[name]
	if !exists {
		return nil, fmt.Errorf("domain not found: %s", name)
	}

	return domain, nil
}

// Domains returns all domains
func (p *Platform) Domains() ([]hosting.Domain, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	domains := make([]hosting.Domain, 0, len(p.domains))
	for _, d := range p.domains {
		domains = append(domains, d)
	}
	return domains, nil
}

// Test control methods

// FailNext causes the next operation to fail
func (p *Platform) FailNext(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.shouldFail = true
	p.failMessage = message
}

// Reset clears the failure state
func (p *Platform) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.shouldFail = false
	p.failMessage = ""
}

// Server implements the hosting.Server interface
type Server struct {
	id       string
	name     string
	ip       string
	platform *Platform
	volumes  map[string]*Volume
	domains  map[string]*Domain
	env      map[string]string
	mu       sync.RWMutex

	// Track operations
	execCalls [][]string
	copyCalls []copyCall
	destroyed bool
}

type copyCall struct {
	src, dst string
}

// ID returns the server ID
func (s *Server) ID() string { return s.id }

// IP returns the server IP
func (s *Server) IP() string { return s.ip }

// Name returns the server name
func (s *Server) Name() string { return s.name }

// Mount attaches a volume to the server
func (s *Server) Mount(volume hosting.Volume) (hosting.Volume, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.platform.shouldFail {
		return nil, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	// Create a mock volume
	s.platform.volumeCount++
	v := &Volume{
		id:       fmt.Sprintf("vol_%d", s.platform.volumeCount),
		name:     fmt.Sprintf("volume-%d", s.platform.volumeCount),
		serverID: s.id,
		platform: s.platform,
	}

	s.volumes[v.id] = v
	s.platform.volumes[v.id] = v

	return v, nil
}

// Volumes returns all volumes attached to the server
func (s *Server) Volumes() ([]hosting.Volume, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.platform.shouldFail {
		return nil, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	volumes := make([]hosting.Volume, 0, len(s.volumes))
	for _, v := range s.volumes {
		volumes = append(volumes, v)
	}
	return volumes, nil
}

// Alias creates a domain alias for the server
func (s *Server) Alias(domain hosting.Domain) (hosting.Domain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.platform.shouldFail {
		return nil, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	// Create a mock domain
	s.platform.domainCount++
	d := &Domain{
		id:       fmt.Sprintf("dom_%d", s.platform.domainCount),
		typ:      "A",
		name:     fmt.Sprintf("domain-%d", s.platform.domainCount),
		data:     s.ip,
		serverID: s.id,
		platform: s.platform,
	}

	s.domains[d.id] = d
	s.platform.domains[d.name] = d

	return d, nil
}

// Domains returns all domains pointing to the server
func (s *Server) Domains() ([]hosting.Domain, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.platform.shouldFail {
		return nil, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	domains := make([]hosting.Domain, 0, len(s.domains))
	for _, d := range s.domains {
		domains = append(domains, d)
	}
	return domains, nil
}

// Env sets an environment variable
func (s *Server) Env(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.platform.shouldFail {
		return fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	if s.env == nil {
		s.env = make(map[string]string)
	}
	s.env[key] = value
	return nil
}

// Copy simulates file copying
func (s *Server) Copy(src, dst string) (bytes.Buffer, bytes.Buffer, error) {
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
func (s *Server) Dump(path string, data []byte) (bytes.Buffer, bytes.Buffer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var stdout, stderr bytes.Buffer

	if s.platform.shouldFail {
		return stdout, stderr, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	stdout.WriteString(fmt.Sprintf("Dumped %d bytes to %s\n", len(data), path))

	return stdout, stderr, nil
}

// Exec simulates command execution
func (s *Server) Exec(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var stdout, stderr bytes.Buffer

	if s.platform.shouldFail {
		return stdout, stderr, fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	s.execCalls = append(s.execCalls, args)
	stdout.WriteString(fmt.Sprintf("Mock execution of: %v\n", args))

	return stdout, stderr, nil
}

// Connect simulates an interactive connection
func (s *Server) Connect(stdin io.Reader, stdout io.Writer, stderr io.Writer, args ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.platform.shouldFail {
		return fmt.Errorf("mock error: %s", s.platform.failMessage)
	}

	fmt.Fprintf(stdout, "Connected to mock server %s\n", s.id)
	return nil
}

// Destroy simulates server destruction
func (s *Server) Destroy(ctx context.Context) error {
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

// Volume implements the hosting.Volume interface
type Volume struct {
	id       string
	name     string
	serverID string
	platform *Platform
}

// ID returns the volume ID
func (v *Volume) ID() string { return v.id }

// Name returns the volume name
func (v *Volume) Name() string { return v.name }

// Server returns the server this volume is attached to
func (v *Volume) Server() (hosting.Server, error) {
	if v.serverID == "" {
		return nil, nil
	}
	return v.platform.Server(v.serverID)
}

// Domain implements the hosting.Domain interface
type Domain struct {
	id       string
	typ      string
	name     string
	data     string
	serverID string
	platform *Platform
}

// ID returns the domain ID
func (d *Domain) ID() string { return d.id }

// Type returns the record type
func (d *Domain) Type() string { return d.typ }

// Name returns the domain name
func (d *Domain) Name() string { return d.name }

// Data returns the record data
func (d *Domain) Data() string { return d.data }

// Server returns the server this domain points to
func (d *Domain) Server() (hosting.Server, error) {
	if d.serverID == "" {
		return nil, nil
	}
	return d.platform.Server(d.serverID)
}