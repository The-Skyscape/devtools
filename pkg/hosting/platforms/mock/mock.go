package mock

import (
	"fmt"
	"sync"

	"github.com/The-Skyscape/devtools/pkg/hosting"
)

// Platform provides a mock hosting platform for testing
type Platform struct {
	servers map[string]*hosting.Server
	volumes map[string]*hosting.Volume
	domains map[string]*hosting.Domain
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
		servers: make(map[string]*hosting.Server),
		volumes: make(map[string]*hosting.Volume),
		domains: make(map[string]*hosting.Domain),
	}
}

// Ensure Platform implements hosting.Platform
var _ hosting.Platform = (*Platform)(nil)

// NewServer creates a new mock server
func (p *Platform) NewServer(server *hosting.Server) (*hosting.Server, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	p.serverCount++
	server.Platform = p
	server.ID = fmt.Sprintf("srv_%d", p.serverCount)
	if server.IP == "" {
		server.IP = fmt.Sprintf("10.0.0.%d", p.serverCount)
	}
	if server.Name == "" {
		server.Name = fmt.Sprintf("server-%d", p.serverCount)
	}
	server.Status = "active"

	p.servers[server.ID] = server

	return server, nil
}

// GetServer retrieves a server by ID
func (p *Platform) GetServer(id string) (*hosting.Server, error) {
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

// AllServers returns all servers
func (p *Platform) AllServers() ([]*hosting.Server, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	servers := make([]*hosting.Server, 0, len(p.servers))
	for _, s := range p.servers {
		servers = append(servers, s)
	}
	return servers, nil
}

// DestroyServer destroys a server by ID
func (p *Platform) DestroyServer(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return fmt.Errorf("mock error: %s", p.failMessage)
	}

	if _, exists := p.servers[id]; !exists {
		return fmt.Errorf("server not found: %s", id)
	}

	delete(p.servers, id)
	return nil
}

// NewVolume creates a new volume
func (p *Platform) NewVolume(volume *hosting.Volume) (*hosting.Volume, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	p.volumeCount++
	volume.Platform = p
	volume.ID = fmt.Sprintf("vol_%d", p.volumeCount)
	if volume.Name == "" {
		volume.Name = fmt.Sprintf("volume-%d", p.volumeCount)
	}

	p.volumes[volume.ID] = volume

	return volume, nil
}

// GetVolume retrieves a volume by ID
func (p *Platform) GetVolume(id string) (*hosting.Volume, error) {
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

// AllVolumes returns all volumes
func (p *Platform) AllVolumes() ([]*hosting.Volume, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	volumes := make([]*hosting.Volume, 0, len(p.volumes))
	for _, v := range p.volumes {
		volumes = append(volumes, v)
	}
	return volumes, nil
}

// MountVolume attaches a volume to a server
func (p *Platform) MountVolume(volume *hosting.Volume, server *hosting.Server) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return fmt.Errorf("mock error: %s", p.failMessage)
	}

	// Simulate mounting
	return nil
}

// UnmountVolume detaches a volume from a server
func (p *Platform) UnmountVolume(volume *hosting.Volume, server *hosting.Server) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return fmt.Errorf("mock error: %s", p.failMessage)
	}

	// Simulate unmounting
	return nil
}

// DestroyVolume destroys a volume by ID
func (p *Platform) DestroyVolume(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return fmt.Errorf("mock error: %s", p.failMessage)
	}

	if _, exists := p.volumes[id]; !exists {
		return fmt.Errorf("volume not found: %s", id)
	}

	delete(p.volumes, id)
	return nil
}

// LookupDomain looks up a domain record
func (p *Platform) LookupDomain(domain *hosting.Domain) (*hosting.Domain, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.shouldFail {
		return nil, fmt.Errorf("mock error: %s", p.failMessage)
	}

	key := fmt.Sprintf("%s.%s", domain.Sub, domain.Name)
	existing, exists := p.domains[key]
	if !exists {
		return nil, nil // Not found is not an error
	}

	return existing, nil
}

// AssignDomain assigns a domain to a server
func (p *Platform) AssignDomain(server *hosting.Server, domain *hosting.Domain) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return fmt.Errorf("mock error: %s", p.failMessage)
	}

	p.domainCount++
	domain.Platform = p
	domain.ID = fmt.Sprintf("dom_%d", p.domainCount)
	domain.Data = server.IP

	key := fmt.Sprintf("%s.%s", domain.Sub, domain.Name)
	p.domains[key] = domain

	return nil
}

// DestroyDomain destroys a domain by ID
func (p *Platform) DestroyDomain(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shouldFail {
		return fmt.Errorf("mock error: %s", p.failMessage)
	}

	// Find and delete domain by ID
	for key, domain := range p.domains {
		if domain.ID == id {
			delete(p.domains, key)
			return nil
		}
	}

	return fmt.Errorf("domain not found with ID: %s", id)
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

// GetServerExecCalls returns the exec calls made to a server (for testing)
func (p *Platform) GetServerExecCalls(serverID string) [][]string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	server, exists := p.servers[serverID]
	if !exists || server == nil {
		return nil
	}

	// Mock implementation - just return empty for now
	// In real tests, you might want to track these separately
	return [][]string{}
}

// GetServerCopyCalls returns the copy calls made to a server (for testing)
func (p *Platform) GetServerCopyCalls(serverID string) []struct{ Src, Dst string } {
	p.mu.RLock()
	defer p.mu.RUnlock()

	server, exists := p.servers[serverID]
	if !exists || server == nil {
		return nil
	}

	// Mock implementation - just return empty for now
	return []struct{ Src, Dst string }{}
}

func (p *Platform) GetMountedServer(volume *hosting.Volume) (*hosting.Server, error) {
	serverID := volume.ID

	server, exists := p.servers[serverID]
	if !exists || server == nil {
		return nil, nil
	}

	return server, nil
}
