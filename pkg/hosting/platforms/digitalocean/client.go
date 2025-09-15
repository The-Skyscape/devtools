package digitalocean

import (
	"context"
	"os"
	"strconv"

	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// ApiKey is the global DigitalOcean API key, loaded from DIGITAL_OCEAN_API_KEY environment variable.
// While global variables are generally discouraged, this provides a simple interface for CLI tools
// and follows the pattern of other cloud SDKs.
var ApiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")

// DigitalOceanClient wraps the godo client with our hosting interface.
// It provides a clean abstraction over DigitalOcean's API.
type DigitalOceanClient struct {
	*godo.Client
}

// Connect creates a new DigitalOcean client with the provided API key.
// If no API key is provided, it falls back to the global ApiKey variable.
// This allows both explicit configuration and environment-based configuration.
func Connect(apiKey string) *DigitalOceanClient {
	// Use provided key or fall back to global variable
	if apiKey == "" {
		apiKey = ApiKey
	}

	return &DigitalOceanClient{
		Client: godo.NewClient(oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: apiKey,
			}),
		)),
	}
}

func (client *DigitalOceanClient) Launch(s *Server, opts ...hosting.LaunchOption) (*Server, error) {
	// &Server{
	// 	client: client,
	// 	Name:   name,
	// 	Size:   "s-1vcpu-1gb",
	// 	Region: "sfo2",
	// 	Image:  "docker-20-04",
	// 	Status: "new",
	// }

	s.client = client
	return s, s.Launch(opts...)
}

func (client *DigitalOceanClient) GetServer(id string) (*Server, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	server := &Server{client: client, ID: intID}
	return server, server.load()
}
