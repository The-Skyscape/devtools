package digitalocean

// ServerOption configures DigitalOcean server launch
type ServerOption func(*Server)

// WithName sets the server name
func WithName(name string) ServerOption {
	return func(s *Server) {
		s.Name = name
	}
}

// WithSize sets the droplet size
func WithSize(size string) ServerOption {
	return func(s *Server) {
		s.Size = size
	}
}

// WithRegion sets the datacenter region
func WithRegion(region string) ServerOption {
	return func(s *Server) {
		s.Region = region
	}
}

// WithImage sets the OS image
func WithImage(image string) ServerOption {
	return func(s *Server) {
		s.Image = image
	}
}

// WithProject assigns to a DigitalOcean project
func WithProject(projectID string) ServerOption {
	return func(s *Server) {
		s.Project = projectID
	}
}

// WithSSHKeys adds SSH key fingerprints
func WithSSHKeys(keys ...string) ServerOption {
	return func(s *Server) {
		// Would need to add SSHKeys field to Server or handle differently
		// For now, this would be handled during Launch
	}
}

// WithTags adds droplet tags
func WithTags(tags ...string) ServerOption {
	return func(s *Server) {
		// Would need to add Tags field to Server or handle differently
		// For now, this would be handled during Launch
	}
}

// WithUserData sets cloud-init user data
func WithUserData(data string) ServerOption {
	return func(s *Server) {
		// Would need to add UserData field to Server or handle differently
		// For now, this would be handled during Launch
	}
}