package containers

import (
	"bytes"
	"cmp"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"text/template"

	"github.com/The-Skyscape/devtools/pkg/database"

	"github.com/pkg/errors"
)

// Host represents a docker host where
// commands can be ran, this can be local
// or remote server
type Host interface {
	SetStdin(io.Reader)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	Exec(...string) error
	Dump(path string, data []byte) error
}

// BuildImage builds a Docker image on the given host
func BuildImage(host Host, tag, context string) error {
	if host == nil {
		return errors.New("host not set")
	}

	return host.Exec("docker", "build", "-t", tag, context)
}

//go:embed resources/start-service.sh
var startService string

// IsRunning checks if the service is currently running
func IsRunning(host Host, s *Service) bool {
	s.Host = cmp.Or(s.Host, host)
	return s.IsRunning()
}

// Launch creates a Docker container with the service configuration
func Launch(host Host, s *Service) (err error) {
	s.Host = cmp.Or(s.Host, host)

	if s.Image == "" {
		return errors.New("missing image")
	}

	var tmpl *template.Template
	if tmpl, err = template.New("start-service").Funcs(template.FuncMap{
		"dataDir": database.DataDir,
	}).Parse(startService); err != nil {
		return errors.Wrap(err, "failed to render start command")
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, s); err != nil {
		return errors.Wrap(err, "failed to start service command")
	}

	// Use stdin with bash to handle complex multi-line scripts
	var stdout, stderr bytes.Buffer
	s.SetStdin(&buf)
	s.SetStdout(&stdout)
	s.SetStderr(&stderr)

	// Log the script being executed for debugging
	log.Printf("Launching container %s with image %s", s.Name, s.Image)

	if err := s.Exec("bash"); err != nil {
		// Include stdout and stderr output in error message and logs
		log.Printf("Container launch failed for %s", s.Name)
		if stdoutStr := stdout.String(); stdoutStr != "" {
			log.Printf("Stdout: %s", stdoutStr)
		}
		if stderrStr := stderr.String(); stderrStr != "" {
			log.Printf("Stderr: %s", stderrStr)
		}
		errMsg := err.Error()
		if stderrStr := stderr.String(); stderrStr != "" {
			errMsg = fmt.Sprintf("%s: %s", errMsg, stderrStr)
		}
		return errors.New(errMsg)
	}

	// Log success with output
	if stdoutStr := stdout.String(); stdoutStr != "" {
		log.Printf("Container %s launched successfully: %s", s.Name, strings.TrimSpace(stdoutStr))
	}

	return nil
}

// ListServices returns a list of all services on
// the given host or an error on failure
func ListServices(host Host) (services []*Service, err error) {
	var stdout bytes.Buffer

	host.SetStdout(&stdout)
	err = host.Exec("docker", "ps", "-a", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	for line := range strings.Lines(stdout.String()) {
		var summary struct {
			ID      string
			State   string
			Image   string
			Names   string
			Command string
		}

		if err = json.Unmarshal([]byte(line), &summary); err != nil {
			return nil, errors.Wrap(err, "failed to fetch services")
		}

		services = append(services, &Service{
			Host:    host,
			ID:      summary.ID,
			Status:  summary.State,
			Name:    summary.Names,
			Image:   summary.Image,
			Command: summary.Command,
		})
	}

	return services, nil
}

// GetService returns a specific service by name from the host
func GetService(host Host, name string) (*Service, error) {
	services, err := ListServices(host)
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if service.Name == name {
			return service, nil
		}
	}

	return nil, errors.New("service not found")
}
