package digitalocean

import (
	"context"
	"os"
	"strconv"
	
	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// DigitalOceanClient wraps the godo client with our hosting interface.
// It provides a clean abstraction over DigitalOcean's API.
type DigitalOceanClient struct {
	*godo.Client
	apiKey string // Store API key internally for reconnection if needed
}

// Connect creates a new DigitalOcean client with the provided API key.
// If no API key is provided, it falls back to the DIGITAL_OCEAN_API_KEY environment variable.
// This allows both explicit configuration and environment-based configuration.
func Connect(apiKey string) *DigitalOceanClient {
	// Use provided key or fall back to environment variable
	if apiKey == "" {
		apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}
	
	return &DigitalOceanClient{
		Client: godo.NewClient(oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: apiKey,
			}),
		)),
		apiKey: apiKey,
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
