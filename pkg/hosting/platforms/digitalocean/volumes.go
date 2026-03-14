package digitalocean

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (client *DigitalOceanClient) NewVolume(volume *hosting.Volume) (*hosting.Volume, error) {
	ctx := context.Background()
	volume.Platform = client

	if volume.ID != "" {
		return nil, errors.New("Volume has already been created: " + volume.ID)
	}

	v, _, err := client.Storage.CreateVolume(ctx, &godo.VolumeCreateRequest{
		Region:         volume.Loc,
		Name:           volume.Name,
		Description:    fmt.Sprintf("Persistent storage for workspace %s", volume.Name),
		SizeGigaBytes:  int64(volume.Size),
		FilesystemType: "ext4",
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create volume")
	}

	if client.DefaultProject != "" {
		fmt.Printf("Assigning volume %s to project %s\n", v.ID, client.DefaultProject)
		_, _, err = client.Projects.AssignResources(ctx, client.DefaultProject,
			fmt.Sprintf("do:volume:%s", v.ID))
		if err != nil {
			// Log but don't fail if project assignment fails
			fmt.Printf("⚠️  Warning: Failed to assign volume to project %s: %v\n", client.DefaultProject, err)
		} else {
			fmt.Printf("✅ Successfully assigned volume %s to project %s\n", v.ID, client.DefaultProject)
		}
	}

	// Wait for volume to be available
	time.Sleep(5 * time.Second)

	volume.ID = v.ID
	return volume, nil
}

func (client *DigitalOceanClient) GetVolume(volumeID string) (*hosting.Volume, error) {
	ctx := context.Background()

	volume, _, err := client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get volume")
	}

	return &hosting.Volume{
		Platform: client,
		ID:       volume.ID,
		Loc:      volume.Region.Slug,
		Name:     volume.Name,
		Size:     int(volume.SizeGigaBytes),
	}, nil
}

func (client *DigitalOceanClient) GetVolumeByName(name string) (*hosting.Volume, error) {
	ctx := context.Background()

	volumes, _, err := client.Storage.ListVolumes(ctx, &godo.ListVolumeParams{
		ListOptions: &godo.ListOptions{},
		Name:        name,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to list volumes")
	}

	if len(volumes) == 0 {
		return nil, errors.New("volume not found")
	}

	vol := volumes[0]
	return &hosting.Volume{
		Platform: client,
		ID:       vol.ID,
		Loc:      vol.Region.Slug,
		Name:     vol.Name,
		Size:     int(vol.SizeGigaBytes),
	}, nil

}

func (client *DigitalOceanClient) AllVolumes() ([]*hosting.Volume, error) {
	ctx := context.Background()

	spaces, _, err := client.Storage.ListVolumes(ctx, &godo.ListVolumeParams{
		ListOptions: &godo.ListOptions{},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to list volumes")
	}

	volumes := []*hosting.Volume{}
	for _, space := range spaces {
		volume := &hosting.Volume{
			Platform: client,
			ID:       space.ID,
			Loc:      space.Region.Slug,
			Name:     space.Name,
			Size:     int(space.SizeGigaBytes),
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

func (client *DigitalOceanClient) MountVolume(volume *hosting.Volume, server *hosting.Server) error {
	if volume == nil {
		return errors.New("volume is nil")
	}

	if server == nil {
		return errors.New("server is nil")
	}

	serverID, err := strconv.Atoi(server.ID)
	if err != nil {
		return errors.Wrap(err, "invalid server ID: "+server.ID)
	}

	ctx := context.Background()
	_, _, err = client.StorageActions.Attach(ctx, volume.ID, serverID)
	if err != nil {
		return errors.Wrap(err, "failed to attach volume")
	}

	time.Sleep(5 * time.Second)

	return server.Connect(nil, os.Stdout, os.Stderr, fmt.Sprintf(`
		mkdir -p /mnt/%[2]s
		mountpoint -q /mnt/%[2]s || mount -o defaults,nofail,discard,noatime /dev/disk/by-id/scsi-0DO_Volume_%[1]s /mnt/%[2]s
	`, volume.Name, strings.ReplaceAll(volume.Name, "-", "_")))
}

func (client *DigitalOceanClient) UnmountVolume(volume *hosting.Volume, server *hosting.Server) error {
	if volume == nil {
		return errors.New("volume is nil")
	}

	if server == nil {
		return errors.New("server is nil")
	}

	serverID, err := strconv.Atoi(server.ID)
	if err != nil {
		return errors.Wrap(err, "invalid server ID: "+server.ID)
	}

	ctx := context.Background()
	_, _, err = client.StorageActions.DetachByDropletID(ctx, volume.ID, serverID)
	if err != nil {
		return errors.Wrap(err, "failed to detach volume")
	}

	time.Sleep(5 * time.Second)
	return nil
}

func (client *DigitalOceanClient) DestroyVolume(volumeID string) error {
	ctx := context.Background()
	_, err := client.Storage.DeleteVolume(ctx, volumeID)
	return errors.Wrap(err, "failed to delete volume")
}

func (client *DigitalOceanClient) GetMountedServer(vol *hosting.Volume) (*hosting.Server, error) {
	volume, _, err := client.Storage.GetVolume(context.Background(), vol.ID)
	if err != nil {
		return nil, err
	}

	if len(volume.DropletIDs) == 0 {
		return nil, errors.New("volume is not mounted")
	}

	dropletID := volume.DropletIDs[0]
	droplet, _, err := client.Droplets.Get(context.Background(), dropletID)
	if err != nil {
		return nil, err
	}

	ipAddr, _ := droplet.PublicIPv4()
	privateIP, _ := droplet.PrivateIPv4()
	return &hosting.Server{
		Platform: client,
		ID:       fmt.Sprintf("%d", droplet.ID),
		IP:       ipAddr,
		PrivIP:   privateIP,
		Loc:      droplet.Region.Slug,
		Size:     droplet.Size.Slug,
		Name:     droplet.Name,
	}, nil
}
