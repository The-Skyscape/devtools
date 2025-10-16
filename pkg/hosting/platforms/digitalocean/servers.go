package digitalocean

import (
	"cmp"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (client *DigitalOceanClient) NewServer(server *hosting.Server) (*hosting.Server, error) {
	ctx := context.Background()
	server.Platform = client

	if server.ID != "" {
		return nil, errors.New("Server has already been created: " + server.ID)
	}

	accessKey, err := client.accessKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get access key")
	}

	slug := cmp.Or(client.DefaultImage, "docker-20-04")
	droplet, _, err := client.Droplets.Create(ctx, &godo.DropletCreateRequest{
		Name:    server.Name,
		Region:  server.Loc,
		Size:    server.Size,
		Tags:    server.Tags,
		Image:   godo.DropletCreateImage{Slug: slug},
		SSHKeys: []godo.DropletCreateSSHKey{{Fingerprint: accessKey.Fingerprint}},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create droplet")
	}

	server.ID = fmt.Sprintf("%d", droplet.ID)
	server.Status = droplet.Status

	if client.DefaultProject != "" {
		fmt.Printf("Assigning droplet %d to project %s\n", droplet.ID, client.DefaultProject)
		_, _, err = client.Projects.AssignResources(ctx, client.DefaultProject,
			fmt.Sprintf("do:droplet:%d", droplet.ID))
		if err != nil {
			// Log but don't fail if project assignment fails
			fmt.Printf("⚠️  Warning: Failed to assign droplet to project %s: %v\n", client.DefaultProject, err)
		} else {
			fmt.Printf("✅ Successfully assigned droplet %d to project %s\n", droplet.ID, client.DefaultProject)
		}
	}

	for server.IP == "" {
		time.Sleep(10 * time.Second)
		if server, err = client.GetServer(server.ID); err != nil {
			return nil, errors.Wrap(err, "failed to get droplet")
		}
	}

	// Give the server time to boot up after getting an IP
	// Even though we have waitForSSH later, the server needs initial boot time
	// before SSH service starts listening
	fmt.Printf("Waiting for server to initialize after getting IP...\n")
	time.Sleep(30 * time.Second)

	return server, nil
}

func (client *DigitalOceanClient) GetServer(id string) (*hosting.Server, error) {
	ctx := context.Background()

	intID, err := strconv.Atoi(id)
	if intID == 0 || err != nil {
		return nil, errors.New("missing droplet id from server")
	}

	droplet, _, err := client.Droplets.Get(ctx, intID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get droplet")
	}

	ipAddress, _ := droplet.PublicIPv4()
	privateIP, _ := droplet.PrivateIPv4()
	return &hosting.Server{
		Platform: client,
		ID:       fmt.Sprintf("%d", droplet.ID),
		IP:       ipAddress,
		PrivIP:   privateIP,
		Loc:      droplet.Region.Name,
		Size:     droplet.SizeSlug,
		Name:     droplet.Name,
		Status:   droplet.Status,
		Tags:     droplet.Tags,
	}, nil
}

func (client *DigitalOceanClient) AllServers() ([]*hosting.Server, error) {
	ctx := context.Background()

	droplets, _, err := client.Droplets.List(ctx, &godo.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list droplets")
	}

	servers := make([]*hosting.Server, 0, len(droplets))
	for _, droplet := range droplets {
		ipAddress, _ := droplet.PublicIPv4()
		privateIP, _ := droplet.PrivateIPv4()
		servers = append(servers, &hosting.Server{
			Platform: client,
			ID:       fmt.Sprintf("%d", droplet.ID),
			IP:       ipAddress,
			PrivIP:   privateIP,
			Loc:      droplet.Region.Name,
			Size:     droplet.SizeSlug,
			Name:     droplet.Name,
			Status:   droplet.Status,
			Tags:     droplet.Tags,
		})
	}

	return servers, nil
}

func (client *DigitalOceanClient) DestroyServer(id string) error {
	ctx := context.Background()

	intID, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("missing droplet id from server")
	}

	_, err = client.Droplets.Delete(ctx, intID)
	return errors.Wrap(err, "failed to delete droplet")
}
