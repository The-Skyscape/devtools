package digitalocean

import (
	"cmp"
	"context"
	"os"
	"github.com/The-Skyscape/devtools/pkg/hosting"
	"strconv"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

var ApiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")

type DigitalOceanClient struct {
	*godo.Client
}

func Connect(apiKey string) *DigitalOceanClient {
	return &DigitalOceanClient{godo.NewClient(oauth2.NewClient(
		context.Background(),
		oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: cmp.Or(apiKey, ApiKey),
		}),
	))}
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
