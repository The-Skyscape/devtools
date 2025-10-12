package infra

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/containers"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func WithServer(name string, opts ...ServerOption) CloudOption {
	return CloudOption{
		OnBoot: func(c *CloudProvider, v *godo.VPC) (err error) {
			droplets, _, err := c.Droplets.List(context.Background(), nil)
			if err != nil {
				return errors.Wrap(err, "failed to get droplet")
			}

			var droplet *godo.Droplet
			for _, d := range droplets {
				if d.Name == name {
					droplet = &d
					break
				}
			}

			if droplet == nil {
				_, err = c.Launch(c.vpc.RegionSlug, name, opts...)
			}

			return errors.Wrap(err, "failed to find or create server")
		},
	}
}

func (c *CloudProvider) Launch(region, name string, opts ...ServerOption) (*godo.Droplet, error) {
	ctx := context.Background()
	accessKey, err := c.ensureKeys()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ssh access key")
	}

	req := godo.DropletCreateRequest{Name: name, Region: region}
	for _, opt := range opts {
		if opt.OnInit != nil {
			if err := opt.OnInit(c, &req); err != nil {
				return nil, errors.Wrap(err, "failed to initialize droplet")
			}
		}
	}

	req.SSHKeys = []godo.DropletCreateSSHKey{{
		Fingerprint: accessKey.Fingerprint,
	}}

	if c.vpc != nil {
		req.Region = c.vpc.RegionSlug
		req.VPCUUID = c.vpc.ID
	}

	droplet, _, err := c.Client.Droplets.Create(ctx, &req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create droplet")
	}

	if c.projectID != "" {
		resourceID := fmt.Sprintf("do:droplet:%d", droplet.ID)
		c.Projects.AssignResources(ctx, c.projectID, resourceID)
	}

	for _, opt := range opts {
		if opt.OnBoot != nil {
			if err := opt.OnBoot(c, droplet); err != nil {
				return nil, errors.Wrap(err, "failed to boot droplet")
			}
		}
	}

	return droplet, nil
}

type ServerOption struct {
	OnInit func(*CloudProvider, *godo.DropletCreateRequest) error
	OnBoot func(*CloudProvider, *godo.Droplet) error
}

func WithSize(size string) ServerOption {
	return ServerOption{
		OnInit: func(c *CloudProvider, s *godo.DropletCreateRequest) error {
			log.Printf("Setting server size: %s", size)
			s.Size = size
			return nil
		},
	}
}

func WithImage(image string) ServerOption {
	return ServerOption{
		OnInit: func(c *CloudProvider, s *godo.DropletCreateRequest) error {
			log.Printf("Setting server image: %s", image)
			s.Image = godo.DropletCreateImage{Slug: image}
			return nil
		},
	}
}

func WithTags(tags []string) ServerOption {
	return ServerOption{
		OnInit: func(c *CloudProvider, s *godo.DropletCreateRequest) error {
			log.Printf("Setting server tags: %v", tags)
			s.Tags = tags
			return nil
		},
	}
}

func WithVolume(name string, size int64) ServerOption {
	return ServerOption{
		OnBoot: func(c *CloudProvider, s *godo.Droplet) error {
			ctx := context.Background()
			vs, _, err := c.Storage.ListVolumes(ctx, nil)
			if err != nil {
				return errors.Wrap(err, "failed to list volumes")
			}

			var v *godo.Volume
			for _, vol := range vs {
				if vol.Name == name {
					v = &vol
					break
				}
			}

			if v == nil {
				log.Printf("Creating volume: %s (%dGB)", name, size)
				v, _, err = c.Storage.CreateVolume(ctx, &godo.VolumeCreateRequest{
					Name:           name,
					Region:         s.Region.Slug,
					SizeGigaBytes:  size,
					FilesystemType: "ext4",
					Description:    "Skyscape HQ centralized data volume",
				})
			} else {
				log.Printf("Found existing volume: %s", name)
			}

			if err != nil {
				return errors.Wrap(err, "failed to find or create volume")
			}

			if c.projectID != "" {
				resourceID := fmt.Sprintf("do:volume:%s", v.ID)
				c.Projects.AssignResources(ctx, c.projectID, resourceID)
			}

			log.Printf("Attaching volume %s to server", name)
			act, _, err := c.StorageActions.Attach(ctx, v.ID, s.ID)
			if err != nil {
				return errors.Wrap(err, "failed to attach volume")
			}

			for {
				a, _, err := c.Actions.Get(ctx, act.ID)
				if err != nil {
					return errors.Wrap(err, "failed to check attachment")
				}
				if a.Status == "completed" {
					break
				}
				time.Sleep(2 * time.Second)
			}

			return nil
		},
	}
}

func WithCopy(src, dst string) ServerOption {
	return ServerOption{
		OnBoot: func(c *CloudProvider, s *godo.Droplet) error {
			log.Printf("Copying %s to server at %s", src, dst)
			ip, err := s.PublicIPv4()
			if err != nil {
				return errors.Wrap(err, "failed to get private ip")
			}

			cmd := exec.Command("scp", "-o",
				"StrictHostKeyChecking=accept-new", src,
				fmt.Sprintf("root@%s:%s", ip, dst))

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return errors.Wrapf(cmd.Run(), "failed to copy %s to %s", src, dst)
		},
	}
}

func WithSetup(script string) ServerOption {
	return ServerOption{
		OnBoot: func(c *CloudProvider, s *godo.Droplet) error {
			log.Printf("Running setup script on server")
			ip, err := s.PublicIPv4()
			if err != nil {
				return errors.Wrap(err, "failed to get public ip")
			}

			cmd := exec.Command("ssh", "-o",
				"StrictHostKeyChecking=accept-new",
				fmt.Sprintf("root@%s", ip), script)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return errors.Wrap(cmd.Run(), "failed to run setup script")
		},
	}
}

type Wrapper struct{ *godo.Droplet }

func (w *Wrapper) SetStdout(io.Writer) {}
func (w *Wrapper) SetStderr(io.Writer) {}
func (w *Wrapper) SetStdin(io.Reader)  {}
func (w *Wrapper) Exec(args ...string) error {
	ip, err := w.PublicIPv4()
	if err != nil {
		return errors.Wrap(err, "failed to get public ip")
	}

	cmd := exec.Command("ssh", "-o",
		"StrictHostKeyChecking=accept-new",
		fmt.Sprintf("root@%s", ip), strings.Join(args, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run(), "failed to execute command")
}

func WithService(service *containers.Service) ServerOption {
	return ServerOption{
		OnBoot: func(c *CloudProvider, s *godo.Droplet) error {
			host := &Wrapper{s}

			if containers.IsRunning(host, service) {
				log.Println("Service is already running")
				return nil
			}

			return containers.Launch(host, service)
		},
	}
}
