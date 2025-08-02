package containers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Service represents a docker container
// instance existing on a given Platform
type Service struct {
	Host

	ID         string
	Status     string
	Name       string
	Image      string
	Network    string
	Privileged bool
	Entrypoint string
	Command    string
	Ports      map[int]int
	Mounts     map[string]string
	Copied     map[string]string
	Env        map[string]string
}

// Stop stops and removes the Docker container
func (s *Service) Stop() error {
	if s.Host == nil {
		return errors.New("platform not set")
	}

	return s.Exec("docker", "stop", s.Name)
}

// Start starts an existing Docker container
func (s *Service) Start() error {
	if s.Host == nil {
		return errors.New("platform not set")
	}

	var stderr bytes.Buffer
	s.SetStderr(&stderr)
	if err := s.Exec("docker", "start", s.Name); err != nil {
		return errors.Wrap(err, stderr.String())
	}

	return nil
}

func (s *Service) Remove() error {
	if s.Host == nil {
		return errors.New("platform not set")
	}

	return s.Exec("docker", "rm", "-f", s.Name)
}

// IsRunning checks if the service is currently running
func (s *Service) IsRunning() bool {
	if s.Host == nil {
		return false
	}

	var stdout bytes.Buffer
	s.SetStdout(&stdout)
	err := s.Exec("docker", "inspect", "-f", "{{.State.Status}}", s.Name)
	return err == nil && strings.TrimSpace(stdout.String()) == "running"
}

// Copy copies a file from the host to the running container
func (s *Service) Copy(srcPath, destPath string) error {
	if s.Host == nil {
		return errors.New("platform not set")
	}

	return s.Exec("docker", "cp", srcPath, s.Name+":"+destPath)
}

func (s *Service) Proxy(port int) http.Handler {
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		log.Fatal("Failed to create reverse proxy: ", err)
	}

	return httputil.NewSingleHostReverseProxy(url)
}
