package containers

import (
	"io"

	"github.com/The-Skyscape/devtools/pkg/hosting"

	"github.com/pkg/errors"
)

func Remote(s hosting.Server) *RemoteHost {
	return &RemoteHost{server: s}
}

type RemoteHost struct {
	server hosting.Server
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (r *RemoteHost) SetStdin(stdin io.Reader)   { r.stdin = stdin }
func (r *RemoteHost) SetStdout(stdout io.Writer) { r.stdout = stdout }
func (r *RemoteHost) SetStderr(stderr io.Writer) { r.stderr = stderr }

func (r *RemoteHost) Exec(args ...string) (err error) {
	return r.server.Connect(r.stdin, r.stdout, r.stderr, args...)
}

func (r *RemoteHost) Service(name string) *Service {
	if s, err := GetService(r, name); err == nil {
		return s
	}

	return &Service{Host: r, Name: name}
}

func (r *RemoteHost) Services() ([]*Service, error) {
	return ListServices(r)
}

func (r *RemoteHost) Launch(service *Service) error {
	if service == nil {
		return errors.New("service configuration is nil")
	}

	if err := Launch(r, service); err != nil {
		return err
	}

	return service.Start()
}

// BuildImage builds a Docker image on the remote platform
func (r *RemoteHost) BuildImage(tag, context string) error {
	return BuildImage(r, tag, context)
}
