package containers

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
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

// Launch creates a Docker container with the service configuration
func Launch(host Host, s *Service) (err error) {
	s.Host = host

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
	var stderr bytes.Buffer
	s.SetStdin(&buf)
	s.SetStderr(&stderr)
	
	if err := s.Exec("bash"); err != nil {
		// Include stderr output in error message
		errMsg := err.Error()
		if stderrStr := stderr.String(); stderrStr != "" {
			errMsg = fmt.Sprintf("%s: %s", errMsg, stderrStr)
		}
		return errors.New(errMsg)
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
