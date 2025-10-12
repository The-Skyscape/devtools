package infra

import (
	"context"
	"log"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type CloudProvider struct {
	*godo.Client
	vpc       *godo.VPC
	projectID string
}

func Bootstrap(apiKey, project, name string, opts ...CloudOption) (*CloudProvider, error) {
	c := CloudProvider{
		Client: godo.NewClient(oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey}),
		)),
		projectID: project,
	}

	vpcs, _, err := c.Client.VPCs.List(context.Background(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup VPCs")
	}

	for _, v := range vpcs {
		if v.Name == name {
			c.vpc = v
			break
		}
	}

	if c.vpc == nil {
		req := godo.VPCCreateRequest{Name: name}
		for _, opt := range opts {
			if opt.OnInit != nil {
				if err := opt.OnInit(&c, &req); err != nil {
					return nil, errors.Wrap(err, "failed to initialize VPC")
				}
			}
		}

		c.vpc, _, err = c.Client.VPCs.Create(context.Background(), &req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create VPC")
		}
	}

	for _, opt := range opts {
		if opt.OnBoot != nil {
			if err := opt.OnBoot(&c, c.vpc); err != nil {
				return nil, errors.Wrap(err, "failed to boot VPC")
			}
		}
	}

	return &c, nil
}

type CloudOption struct {
	OnInit func(*CloudProvider, *godo.VPCCreateRequest) error
	OnBoot func(*CloudProvider, *godo.VPC) error
}

func WithRegion(region string) CloudOption {
	return CloudOption{
		OnInit: func(c *CloudProvider, r *godo.VPCCreateRequest) error {
			log.Printf("Setting VPC region: %s", region)
			r.RegionSlug = region
			return nil
		},
	}
}

func WithIpRange(cidr string) CloudOption {
	return CloudOption{
		OnInit: func(c *CloudProvider, r *godo.VPCCreateRequest) error {
			log.Printf("Setting VPC IP range: %s", cidr)
			r.IPRange = cidr
			return nil
		},
	}
}
