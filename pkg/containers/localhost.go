package containers

import (
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

func Local() *LocalHost {
	return &LocalHost{}
}

type LocalHost struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (l *LocalHost) SetStdin(stdin io.Reader)   { l.stdin = stdin }
func (l *LocalHost) SetStdout(stdout io.Writer) { l.stdout = stdout }
func (l *LocalHost) SetStderr(stderr io.Writer) { l.stderr = stderr }

func (l *LocalHost) Exec(args ...string) (err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = l.stdin
	cmd.Stdout = l.stdout
	cmd.Stderr = l.stderr
	return cmd.Run()
}

func (l *LocalHost) Service(name string) *Service {
	if s, err := GetService(l, name); err == nil {
		return s
	}

	return &Service{Host: l, Name: name}
}

func (l *LocalHost) Services() ([]*Service, error) {
	return ListServices(l)
}

func (l *LocalHost) Launch(service *Service) error {
	if service == nil {
		return errors.New("service configuration is nil")
	}

	if err := Launch(l, service); err != nil {
		return err
	}

	return service.Start()
}

// BuildImage builds a Docker image on the local platform
func (l *LocalHost) BuildImage(tag, context string) error {
	return BuildImage(l, tag, context)
}
