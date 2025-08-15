package containers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Service represents a docker container
// instance existing on a given Platform
type Service struct {
	Host

	ID           string
	Status       string
	Name         string
	Image        string
	Network      string
	Privileged   bool
	Entrypoint   string
	Command      string
	RestartPolicy string // e.g., "always", "unless-stopped", "on-failure"
	Ports        map[int]int
	Mounts       map[string]string
	Copied       map[string]string
	Env          map[string]string
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

// ExecInContainer executes a command inside the running container
func (s *Service) ExecInContainer(command ...string) error {
	if s.Host == nil {
		return errors.New("host not set")
	}

	if !s.IsRunning() {
		return errors.New("container is not running")
	}

	// Build docker exec command
	args := append([]string{"docker", "exec", s.Name}, command...)
	return s.Host.Exec(args...)
}

// ExecInContainerWithOutput executes a command and returns its output
func (s *Service) ExecInContainerWithOutput(command ...string) (string, error) {
	if s.Host == nil {
		return "", errors.New("host not set")
	}

	if !s.IsRunning() {
		return "", errors.New("container is not running")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	s.SetStdout(&stdout)
	s.SetStderr(&stderr)

	// Build docker exec command
	args := append([]string{"docker", "exec", s.Name}, command...)
	if err := s.Host.Exec(args...); err != nil {
		return "", errors.Wrapf(err, "exec failed: %s", stderr.String())
	}

	return stdout.String(), nil
}

// Restart restarts the container
func (s *Service) Restart() error {
	if s.Host == nil {
		return errors.New("host not set")
	}

	var stderr bytes.Buffer
	s.SetStderr(&stderr)
	if err := s.Host.Exec("docker", "restart", s.Name); err != nil {
		return errors.Wrapf(err, "restart failed: %s", stderr.String())
	}

	return nil
}

// GetLogs retrieves container logs
func (s *Service) GetLogs(tail int) (string, error) {
	if s.Host == nil {
		return "", errors.New("host not set")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	s.SetStdout(&stdout)
	s.SetStderr(&stderr)

	args := []string{"docker", "logs", s.Name}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	if err := s.Host.Exec(args...); err != nil {
		return "", errors.Wrapf(err, "failed to get logs: %s", stderr.String())
	}

	return stdout.String(), nil
}

// WaitForReady waits for container to be ready with optional health check
func (s *Service) WaitForReady(timeout time.Duration, healthCheck func() error) error {
	if s.Host == nil {
		return errors.New("host not set")
	}

	start := time.Now()
	for {
		if s.IsRunning() {
			// If health check provided, use it
			if healthCheck != nil {
				if err := healthCheck(); err == nil {
					return nil
				}
			} else {
				// No health check, just check if running
				return nil
			}
		}

		if time.Since(start) > timeout {
			return errors.New("timeout waiting for container to be ready")
		}

		time.Sleep(500 * time.Millisecond)
	}
}
