package hosting

// Platform defines the minimal interface for hosting platforms
type Platform interface {
	// Server operations
	NewServer(server *Server) (*Server, error)
	GetServer(id string) (*Server, error)
	AllServers() ([]*Server, error)
	DestroyServer(id string) error

	// Volume operations
	NewVolume(volume *Volume) (*Volume, error)
	GetVolume(id string) (*Volume, error)
	AllVolumes() ([]*Volume, error)
	MountVolume(volume *Volume, server *Server) error
	UnmountVolume(volume *Volume, server *Server) error
	DestroyVolume(id string) error

	// Domain operations
	LookupDomain(domain *Domain) (*Domain, error)
	AssignDomain(server *Server, domain *Domain) error
	DestroyDomain(id string) error
}
