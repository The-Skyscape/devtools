package mock

// ServerOption configures mock server launch
type ServerOption func(*Server)

// WithName sets the server name
func WithName(name string) ServerOption {
	return func(s *Server) {
		s.name = name
	}
}

// WithIP sets the server IP
func WithIP(ip string) ServerOption {
	return func(s *Server) {
		s.ip = ip
	}
}