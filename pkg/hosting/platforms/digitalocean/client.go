package digitalocean

import (
	"cmp"
	"context"
	"os"

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
	DefaultProject string // Default project ID for new resources
	DefaultImage   string // Default image for new resources
}

var _ hosting.Platform = &DigitalOceanClient{}

// Connect creates a new DigitalOcean client with the provided API key.
// If no API key is provided, it falls back to the global ApiKey variable.
// This allows both explicit configuration and environment-based configuration.
func Connect(apiKey string, opts ...ClientOption) *DigitalOceanClient {
	client := &DigitalOceanClient{
		Client: godo.NewClient(oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: cmp.Or(apiKey, ApiKey),
			}),
		)),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type ClientOption func(client *DigitalOceanClient)

func WithProjectID(projectID string) ClientOption {
	return func(client *DigitalOceanClient) {
		client.DefaultProject = projectID
	}
}

func WithImage(image string) ClientOption {
	return func(client *DigitalOceanClient) {
		client.DefaultImage = image
	}
}
