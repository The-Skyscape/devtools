package containers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Service represents a docker container
// instance existing on a given Platform
type Service struct {
	Host

	ID            string
	Status        string
	Name          string
	Image         string
	Network       string
	Entrypoint    string
	Command       string
	RestartPolicy string // e.g., "always", "unless-stopped", "on-failure"
	Privileged    bool
	Ports         map[int]int
	Mounts        map[string]string
	Copied        map[string]string
	Env           map[string]string

	// Resource limits for container security
	MemoryLimit  string   // e.g., "512m", "2g"
	CPULimit     string   // e.g., "0.5", "2.0" (number of CPUs)
	PidsLimit    int      // Maximum number of PIDs (0 = unlimited)
	ReadOnly     bool     // Make root filesystem read-only
	SecurityOpts []string // Security options like "no-new-privileges"
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
	name := s.Name
	if s.Network == "host" {
		name = "localhost"
	}

	targetURL, err := url.Parse(fmt.Sprintf("http://%s:%d", name, port))
	if err != nil {
		// This should never happen with a valid port number
		// Return a handler that returns an error response
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, fmt.Sprintf("Invalid proxy configuration: %v", err), http.StatusInternalServerError)
		})
	}

	return httputil.NewSingleHostReverseProxy(targetURL)
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

// WithSecurityDefaults applies secure default settings to the container
// This includes memory limits, CPU limits, and security options
func (s *Service) WithSecurityDefaults() *Service {
	// Set reasonable resource limits if not already set
	if s.MemoryLimit == "" {
		s.MemoryLimit = "512m" // Default 512MB memory limit
	}
	if s.CPULimit == "" {
		s.CPULimit = "1.0" // Default 1 CPU limit
	}
	if s.PidsLimit == 0 {
		s.PidsLimit = 256 // Default PID limit to prevent fork bombs
	}

	// Add security options if not already present
	hasNoNewPrivileges := false
	for _, opt := range s.SecurityOpts {
		if opt == "no-new-privileges" {
			hasNoNewPrivileges = true
			break
		}
	}
	if !hasNoNewPrivileges {
		s.SecurityOpts = append(s.SecurityOpts, "no-new-privileges")
	}

	return s
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

// Stats retrieves container statistics
func (s *Service) Stats() (*ContainerStats, error) {
	if s.Host == nil {
		return nil, errors.New("host not set")
	}

	if !s.IsRunning() {
		return nil, errors.New("container is not running")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	s.SetStdout(&stdout)
	s.SetStderr(&stderr)

	// Get container stats in a parseable format
	args := []string{"docker", "stats", s.Name, "--no-stream", "--format",
		"{{.Container}},{{.CPUPerc}},{{.MemUsage}},{{.MemPerc}},{{.NetIO}},{{.BlockIO}}"}

	if err := s.Host.Exec(args...); err != nil {
		return nil, errors.Wrapf(err, "failed to get stats: %s", stderr.String())
	}

	// Parse the stats output
	stats := &ContainerStats{
		Name: s.Name,
		ID:   s.ID,
	}

	// Basic parsing - can be enhanced
	output := strings.TrimSpace(stdout.String())
	if output != "" {
		parts := strings.Split(output, ",")
		if len(parts) >= 4 {
			// Parse CPU percentage
			cpuStr := strings.TrimSuffix(parts[1], "%")
			if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
				stats.CPUPercent = cpu
			}

			// Parse memory percentage
			memStr := strings.TrimSuffix(parts[3], "%")
			if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
				stats.MemPercent = mem
			}

			// Store raw network and block I/O strings
			if len(parts) >= 5 {
				stats.NetIO = parts[4]
			}
			if len(parts) >= 6 {
				stats.BlockIO = parts[5]
			}
		}
	}

	return stats, nil
}

// HealthCheck performs a health check on the container
func (s *Service) HealthCheck() error {
	if !s.IsRunning() {
		return errors.New("container is not running")
	}

	// Basic health check - just verify container is running
	// Can be overridden by specific services for custom health checks
	return nil
}

// RestartIfUnhealthy restarts the container if health check fails
func (s *Service) RestartIfUnhealthy() error {
	if err := s.HealthCheck(); err != nil {
		log.Printf("Health check failed for %s: %v, restarting...", s.Name, err)
		return s.Restart()
	}
	return nil
}
