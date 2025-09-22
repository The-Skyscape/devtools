package digitalocean

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/digitalocean/godo"
)

// Platform wraps DigitalOceanClient to implement hosting.Platform
type Platform struct {
	client *DigitalOceanClient
}

// NewPlatform creates a new DigitalOcean platform
func NewPlatform(apiKey string) *Platform {
	return &Platform{
		client: Connect(apiKey),
	}
}

// NewPlatformWithProject creates a new DigitalOcean platform with project support
func NewPlatformWithProject(apiKey, projectID string) *Platform {
	return &Platform{
		client: ConnectWithProject(apiKey, projectID),
	}
}

// Launch creates a new server
func (p *Platform) Launch(opts ...interface{}) (hosting.Server, error) {
	// Create server with defaults
	doServer := &Server{
		client:  p.client,
		Name:    fmt.Sprintf("server-%d", time.Now().Unix()),
		Size:    "s-2vcpu-2gb",
		Region:  "sfo3",
		Image:   "docker-20-04",
	}

	// Apply options directly to the server
	for _, opt := range opts {
		if fn, ok := opt.(ServerOption); ok {
			fn(doServer)
		}
	}

	// Launch the server
	if err := doServer.Launch(); err != nil {
		return nil, err
	}

	// Wrap in adapter
	adapter := &ServerAdapter{
		server:   doServer,
		platform: p,
	}

	return adapter, nil
}

// Server retrieves a server by ID
func (p *Platform) Server(id string) (hosting.Server, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	doServer, err := p.client.GetServer(strconv.Itoa(intID))
	if err != nil {
		return nil, err
	}

	adapter := &ServerAdapter{
		server:   doServer,
		platform: p,
	}

	return adapter, nil
}

// Servers returns all servers
func (p *Platform) Servers() ([]hosting.Server, error) {
	ctx := context.Background()

	droplets, _, err := p.client.Droplets.List(ctx, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	servers := make([]hosting.Server, 0, len(droplets))
	for _, droplet := range droplets {
		doServer := &Server{
			client: p.client,
			ID:     droplet.ID,
			Name:   droplet.Name,
			Status: droplet.Status,
			IP:     droplet.Networks.V4[0].IPAddress,
		}

		adapter := &ServerAdapter{
			server:   doServer,
			platform: p,
		}

		servers = append(servers, adapter)
	}

	return servers, nil
}

// Volume retrieves a volume by ID
func (p *Platform) Volume(id string) (hosting.Volume, error) {
	volume, err := p.client.GetVolume(id)
	if err != nil {
		return nil, err
	}

	adapter := &VolumeAdapter{
		volume:   volume,
		platform: p,
	}

	return adapter, nil
}

// Volumes returns all volumes
func (p *Platform) Volumes() ([]hosting.Volume, error) {
	ctx := context.Background()

	volumes, _, err := p.client.Storage.ListVolumes(ctx, &godo.ListVolumeParams{})
	if err != nil {
		return nil, err
	}

	result := make([]hosting.Volume, 0, len(volumes))
	for _, vol := range volumes {
		adapter := &VolumeAdapter{
			volume:   &vol,
			platform: p,
		}

		result = append(result, adapter)
	}

	return result, nil
}

// Domain retrieves a domain by name
func (p *Platform) Domain(name string) (hosting.Domain, error) {
	ctx := context.Background()

	// Parse subdomain and base domain
	// For simplicity, assume format is "subdomain.domain.tld"
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid domain format: %s", name)
	}

	subdomain := parts[0]
	baseDomain := strings.Join(parts[1:], ".")

	records, _, err := p.client.Domains.Records(ctx, baseDomain, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.Name == subdomain {
			adapter := &DomainAdapter{
				record:   &record,
				domain:   baseDomain,
				platform: p,
			}

			return adapter, nil
		}
	}

	return nil, fmt.Errorf("domain not found: %s", name)
}

// Domains returns all domains
func (p *Platform) Domains() ([]hosting.Domain, error) {
	ctx := context.Background()

	domains, _, err := p.client.Domains.List(ctx, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]hosting.Domain, 0)
	for _, domain := range domains {
		records, _, err := p.client.Domains.Records(ctx, domain.Name, &godo.ListOptions{})
		if err != nil {
			continue
		}

		for _, record := range records {
			adapter := &DomainAdapter{
				record:   &record,
				domain:   domain.Name,
				platform: p,
			}

			result = append(result, adapter)
		}
	}

	return result, nil
}

// ServerAdapter wraps a DigitalOcean Server to implement hosting.Server
type ServerAdapter struct {
	server   *Server
	platform *Platform
}

// ID returns the server ID
func (a *ServerAdapter) ID() string {
	return strconv.Itoa(a.server.ID)
}

// IP returns the server IP
func (a *ServerAdapter) IP() string {
	return a.server.IP
}

// Name returns the server name
func (a *ServerAdapter) Name() string {
	return a.server.Name
}

// Mount attaches a volume to the server
func (a *ServerAdapter) Mount(volume hosting.Volume) (hosting.Volume, error) {
	if volume == nil {
		return nil, fmt.Errorf("volume is nil")
	}

	volumeID := volume.ID()
	err := a.platform.client.AttachVolume(volumeID, a.server.ID)
	if err != nil {
		return nil, err
	}

	return volume, nil
}

// Volumes returns all volumes attached to the server
func (a *ServerAdapter) Volumes() ([]hosting.Volume, error) {
	ctx := context.Background()

	volumes, _, err := a.platform.client.Storage.ListVolumes(ctx, &godo.ListVolumeParams{})
	if err != nil {
		return nil, err
	}

	result := make([]hosting.Volume, 0)
	for _, vol := range volumes {
		// Check if volume is attached to this server
		for _, dropletID := range vol.DropletIDs {
			if dropletID == a.server.ID {
				adapter := &VolumeAdapter{
					volume:   &vol,
					platform: a.platform,
				}

				result = append(result, adapter)
				break
			}
		}
	}

	return result, nil
}

// Alias creates a DNS alias for the server
func (a *ServerAdapter) Alias(domain hosting.Domain) (hosting.Domain, error) {
	// Extract subdomain and domain from the provided domain
	if domain == nil {
		return nil, fmt.Errorf("domain is nil")
	}

	name := domain.Name()
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid domain format: %s", name)
	}

	subdomain := parts[0]
	baseDomain := strings.Join(parts[1:], ".")

	err := a.server.Alias(subdomain, baseDomain)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

// Domains returns all domains pointing to the server
func (a *ServerAdapter) Domains() ([]hosting.Domain, error) {
	ctx := context.Background()

	domains, _, err := a.platform.client.Domains.List(ctx, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]hosting.Domain, 0)
	for _, domain := range domains {
		records, _, err := a.platform.client.Domains.Records(ctx, domain.Name, &godo.ListOptions{})
		if err != nil {
			continue
		}

		for _, record := range records {
			if record.Type == "A" && record.Data == a.server.IP {
				adapter := &DomainAdapter{
					record:   &record,
					domain:   domain.Name,
					platform: a.platform,
				}

				result = append(result, adapter)
			}
		}
	}

	return result, nil
}

// Env sets an environment variable
func (a *ServerAdapter) Env(key, value string) error {
	return a.server.Env(key, value)
}

// Copy copies a file to the server
func (a *ServerAdapter) Copy(src, dst string) (bytes.Buffer, bytes.Buffer, error) {
	return a.server.Copy(src, dst)
}

// Dump writes data to a file on the server
func (a *ServerAdapter) Dump(path string, data []byte) (bytes.Buffer, bytes.Buffer, error) {
	return a.server.Dump(path, data)
}

// Exec executes a command on the server
func (a *ServerAdapter) Exec(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return a.server.Exec(args...)
}

// Connect establishes an interactive connection
func (a *ServerAdapter) Connect(stdin io.Reader, stdout io.Writer, stderr io.Writer, args ...string) error {
	return a.server.Connect(stdin, stdout, stderr, args...)
}

// Destroy destroys the server
func (a *ServerAdapter) Destroy(ctx context.Context) error {
	return a.server.Destroy(ctx)
}

// VolumeAdapter wraps a godo.Volume to implement hosting.Volume
type VolumeAdapter struct {
	volume   *godo.Volume
	platform *Platform
}

// ID returns the volume ID
func (a *VolumeAdapter) ID() string {
	return a.volume.ID
}

// Name returns the volume name
func (a *VolumeAdapter) Name() string {
	return a.volume.Name
}

// Server returns the server this volume is attached to
func (a *VolumeAdapter) Server() (hosting.Server, error) {
	if len(a.volume.DropletIDs) == 0 {
		return nil, nil
	}

	return a.platform.Server(strconv.Itoa(a.volume.DropletIDs[0]))
}

// DomainAdapter wraps a godo.DomainRecord to implement hosting.Domain
type DomainAdapter struct {
	record   *godo.DomainRecord
	domain   string
	platform *Platform
}

// ID returns the domain record ID
func (a *DomainAdapter) ID() string {
	return strconv.Itoa(a.record.ID)
}

// Type returns the record type
func (a *DomainAdapter) Type() string {
	return a.record.Type
}

// Name returns the full domain name
func (a *DomainAdapter) Name() string {
	if a.record.Name == "@" {
		return a.domain
	}
	return fmt.Sprintf("%s.%s", a.record.Name, a.domain)
}

// Data returns the record data
func (a *DomainAdapter) Data() string {
	return a.record.Data
}

// Server returns the server this domain points to
func (a *DomainAdapter) Server() (hosting.Server, error) {
	if a.record.Type != "A" {
		return nil, nil
	}

	// Find server by IP
	servers, err := a.platform.Servers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.IP() == a.record.Data {
			return server, nil
		}
	}

	return nil, nil
}