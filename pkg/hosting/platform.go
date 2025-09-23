package hosting

// Platform defines the minimal interface for hosting platforms
type Platform interface {
	// Server operations
	NewLaunch(server *Server) (*Server, error)
	GetServer(id string) (*Server, error)
	AllServers() ([]*Server, error)
	DestroyServer(id string) error

	// Volume operations
	NewVolume(volume *Volume) (*Volume, error)
	GetVolume(id string) (*Volume, error)
	AllVolumes() ([]*Volume, error)
	MountVolume(volume *Volume, server *Server) error
	DestroyVolume(id string) error

	// Domain operations
	NewDomain(domain *Domain) (*Domain, error)
	GetDomain(name string) (*Domain, error)
	AllDomains() ([]*Domain, error)
	AssignDomain(domain *Domain, server *Server) error
	DestroyDomain(name string) error
}
