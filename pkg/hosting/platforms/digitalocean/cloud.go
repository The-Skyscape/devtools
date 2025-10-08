package digitalocean

import (
	"cmp"
	"context"
	"fmt"
	"log"

	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

type Cloud struct {
	*godo.VPC
	client *DigitalOceanClient
}

func (client *DigitalOceanClient) Cloud(region, name string) *Cloud {
	ctx := context.Background()
	vpc, _, err := client.VPCs.Get(ctx, name)
	if err != nil {
		vpc, _, err = client.VPCs.Create(ctx, &godo.VPCCreateRequest{
			Name:       name,
			RegionSlug: region,
		})

		if err != nil {
			log.Fatal("Failed to create VPC: ", err)
		}
	}

	return &Cloud{VPC: vpc, client: client}
}

type CloudOpt func(cloud *Cloud) error

func (cloud *Cloud) Ensure(opts ...CloudOpt) {
	for _, opt := range opts {
		if err := opt(cloud); err != nil {
			log.Fatal("Failed to ensure cloud: ", err)
		}
	}
}

type ServerOpt interface {
	OnLaunch(cloud *Cloud, server *hosting.Server) error
	OnSetup(cloud *Cloud, server *hosting.Server) error
}

func WithServer(opts ...ServerOpt) CloudOpt {
	return func(cloud *Cloud) error {
		ctx := context.Background()

		server := &hosting.Server{
			Platform: cloud.client,
			Name:     fmt.Sprintf("skyscape-%s", database.RandomString(8)),
		}
		for _, opt := range opts {
			if err := opt.OnLaunch(cloud, server); err != nil {
				return errors.Wrap(err, "failed to configure server")
			}
		}

		accessKey, err := cloud.client.accessKey()
		if err != nil {
			return errors.Wrap(err, "failed to get access key")
		}

		image := cmp.Or(cloud.client.DefaultImage, "docker-20-04")
		droplet, _, err := cloud.client.Droplets.Create(ctx, &godo.DropletCreateRequest{
			Name:    server.Name,
			Region:  server.Loc,
			Size:    server.Size,
			Image:   godo.DropletCreateImage{Slug: image},
			SSHKeys: []godo.DropletCreateSSHKey{{ID: accessKey.ID}},
			VPCUUID: cloud.ID,
		})

		if err != nil {
			return errors.Wrap(err, "failed to create droplet")
		}

		server.ID = fmt.Sprintf("%d", droplet.ID)
		server.Status = droplet.Status

		for _, opt := range opts {
			if err := opt.OnSetup(cloud, server); err != nil {
				return errors.Wrap(err, "failed to configure server")
			}
		}

		return nil
	}
}
