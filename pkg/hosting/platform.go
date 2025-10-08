// Package hosting provides a unified interface for managing cloud infrastructure
// across multiple providers. It implements a Platform-Resource pattern where
// platforms create and manage resources, and resources hold a reference back
// to their platform for operations.
//
// The hosting package follows a pattern similar to the standard library's
// database/sql package, where DB creates Rows, and Rows maintains a reference
// to its originating DB.
//
// Basic usage:
//
//	platform := digitalocean.Connect(apiKey)
//	server := &hosting.Server{Name: "web-server", Size: "s-2vcpu-4gb"}
//	server, err := platform.NewServer(server)
//	if err != nil {
//		return err
//	}
//	defer server.Destroy()
//
// The Platform interface is implemented by provider-specific clients such as
// digitalocean.DigitalOceanClient. Resources created by a Platform maintain
// a reference to that Platform, allowing operations like server.Destroy()
// to work without explicitly passing the platform.
package hosting

// Platform defines the interface that all hosting providers must implement.
// A Platform manages the lifecycle of cloud resources including servers,
// volumes, and domains.
//
// Implementations should be safe for concurrent use.
type Platform interface {
	// NewServer creates a new server with the specified configuration.
	// The server parameter should contain the desired configuration.
	// The returned server will have its ID, IP, and Platform fields populated.
	// The Platform implementation must set server.Platform to itself.
	//
	// Example:
	//	server := &Server{Name: "web", Size: "s-2vcpu-4gb", Loc: "nyc3"}
	//	server, err = platform.NewServer(server)
	NewServer(server *Server) (*Server, error)

	// GetServer retrieves an existing server by its ID.
	// Returns an error if the server does not exist.
	// The returned server will have its Platform field set.
	GetServer(id string) (*Server, error)

	// AllServers returns a list of all servers managed by this platform.
	// The returned servers will have their Platform fields set.
	AllServers() ([]*Server, error)

	// DestroyServer permanently deletes a server and its associated resources.
	// This operation cannot be undone.
	DestroyServer(id string) error

	// NewVolume creates a new storage volume with the specified configuration.
	// The volume parameter should contain the desired size, location, and name.
	// The returned volume will have its ID and Platform fields populated.
	// The Platform implementation must set volume.Platform to itself.
	NewVolume(volume *Volume) (*Volume, error)

	// GetVolume retrieves an existing volume by its ID.
	// Returns an error if the volume does not exist.
	// The returned volume will have its Platform field set.
	GetVolume(id string) (*Volume, error)

	// AllVolumes returns a list of all volumes managed by this platform.
	// The returned volumes will have their Platform fields set.
	AllVolumes() ([]*Volume, error)

	// MountVolume attaches a volume to a server.
	// Both the volume and server must exist and be in a valid state.
	// The volume should be in the same location as the server.
	MountVolume(volume *Volume, server *Server) error

	// UnmountVolume detaches a volume from a server.
	// The volume must be currently mounted to the server.
	UnmountVolume(volume *Volume, server *Server) error

	GetMountedServer(volume *Volume) (*Server, error)

	// DestroyVolume permanently deletes a volume.
	// The volume must not be mounted to any server.
	// This operation cannot be undone and all data will be lost.
	DestroyVolume(id string) error

	// LookupDomain searches for an existing domain record.
	// The domain parameter should have Name, Sub, and Type fields set.
	// Returns the domain with its ID and Data fields populated if found.
	// Returns nil, nil if the domain record does not exist.
	LookupDomain(domain *Domain) (*Domain, error)

	// AssignDomain creates a DNS record pointing to the specified server.
	// The domain parameter should have Name, Sub, and Type fields set.
	// The server's IP address will be used as the record's data.
	// The Platform implementation must set domain.Platform to itself.
	AssignDomain(server *Server, domain *Domain) error

	// DestroyDomain removes a DNS record by its ID.
	// This operation cannot be undone.
	DestroyDomain(id string) error
}
