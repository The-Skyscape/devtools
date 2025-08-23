package digitalocean

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// CreateVolume creates a new block storage volume
func (client *DigitalOceanClient) CreateVolume(name string, sizeGB int, region string) (*godo.Volume, error) {
	ctx := context.Background()
	
	createRequest := &godo.VolumeCreateRequest{
		Region:        region,
		Name:          name,
		Description:   fmt.Sprintf("Persistent storage for workspace %s", name),
		SizeGigaBytes: int64(sizeGB),
		FilesystemType: "ext4",
	}
	
	volume, _, err := client.Storage.CreateVolume(ctx, createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}
	
	// Wait for volume to be available
	if err := client.WaitForVolumeReady(volume.ID, 60); err != nil {
		return nil, fmt.Errorf("volume creation timed out: %w", err)
	}
	
	return volume, nil
}

// AttachVolume attaches a volume to a droplet
func (client *DigitalOceanClient) AttachVolume(volumeID string, dropletID int) error {
	ctx := context.Background()
	
	action, _, err := client.StorageActions.Attach(ctx, volumeID, dropletID)
	if err != nil {
		return fmt.Errorf("failed to attach volume: %w", err)
	}
	
	// Wait for action to complete
	if err := client.WaitForAction(action.ID, 60); err != nil {
		return fmt.Errorf("volume attach timed out: %w", err)
	}
	
	return nil
}

// DetachVolume detaches a volume from its droplet
func (client *DigitalOceanClient) DetachVolume(volumeID string) error {
	ctx := context.Background()
	
	// Get the volume to find which droplet it's attached to
	volume, _, err := client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	
	if len(volume.DropletIDs) == 0 {
		// Volume is not attached to any droplet
		return nil
	}
	
	// Detach from the first (and should be only) droplet
	action, _, err := client.StorageActions.DetachByDropletID(ctx, volumeID, volume.DropletIDs[0])
	if err != nil {
		return fmt.Errorf("failed to detach volume: %w", err)
	}
	
	// Wait for action to complete
	if err := client.WaitForAction(action.ID, 60); err != nil {
		return fmt.Errorf("volume detach timed out: %w", err)
	}
	
	return nil
}

// ResizeVolume resizes a volume to a new size
func (client *DigitalOceanClient) ResizeVolume(volumeID string, newSizeGB int) error {
	ctx := context.Background()
	
	// Get the volume to find its region
	volume, _, err := client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	
	if int64(newSizeGB) <= volume.SizeGigaBytes {
		return fmt.Errorf("new size must be larger than current size (%d GB)", volume.SizeGigaBytes)
	}
	
	action, _, err := client.StorageActions.Resize(ctx, volumeID, newSizeGB, volume.Region.Slug)
	if err != nil {
		return fmt.Errorf("failed to resize volume: %w", err)
	}
	
	// Wait for action to complete (resizing can take longer)
	if err := client.WaitForAction(action.ID, 300); err != nil {
		return fmt.Errorf("volume resize timed out: %w", err)
	}
	
	return nil
}

// DeleteVolume deletes a volume
func (client *DigitalOceanClient) DeleteVolume(volumeID string) error {
	ctx := context.Background()
	
	// First ensure it's detached
	if err := client.DetachVolume(volumeID); err != nil {
		return fmt.Errorf("failed to detach before delete: %w", err)
	}
	
	// Small delay to ensure detach is complete
	time.Sleep(2 * time.Second)
	
	_, err := client.Storage.DeleteVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}
	
	return nil
}

// GetVolume retrieves a volume by ID
func (client *DigitalOceanClient) GetVolume(volumeID string) (*godo.Volume, error) {
	ctx := context.Background()
	
	volume, _, err := client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}
	
	return volume, nil
}

// ListVolumes lists all volumes in the account
func (client *DigitalOceanClient) ListVolumes() ([]godo.Volume, error) {
	ctx := context.Background()
	
	volumes := []godo.Volume{}
	opt := &godo.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	
	for {
		volumePage, resp, err := client.Storage.ListVolumes(ctx, &godo.ListVolumeParams{
			ListOptions: opt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list volumes: %w", err)
		}
		
		volumes = append(volumes, volumePage...)
		
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		
		opt.Page++
	}
	
	return volumes, nil
}

// WaitForVolumeReady waits for a volume to be in "available" state
func (client *DigitalOceanClient) WaitForVolumeReady(volumeID string, timeoutSeconds int) error {
	ctx := context.Background()
	start := time.Now()
	timeout := time.Duration(timeoutSeconds) * time.Second
	
	for {
		volume, _, err := client.Storage.GetVolume(ctx, volumeID)
		if err != nil {
			return fmt.Errorf("failed to check volume status: %w", err)
		}
		
		// Volume is ready when it has no pending actions
		if len(volume.DropletIDs) == 0 {
			// Not attached, which means it's available
			return nil
		}
		
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for volume to be ready")
		}
		
		time.Sleep(2 * time.Second)
	}
}

// WaitForAction waits for a DigitalOcean action to complete
func (client *DigitalOceanClient) WaitForAction(actionID int, timeoutSeconds int) error {
	ctx := context.Background()
	start := time.Now()
	timeout := time.Duration(timeoutSeconds) * time.Second
	
	for {
		action, _, err := client.Actions.Get(ctx, actionID)
		if err != nil {
			return fmt.Errorf("failed to check action status: %w", err)
		}
		
		if action.Status == "completed" {
			return nil
		}
		
		if action.Status == "errored" {
			return fmt.Errorf("action failed: %s", action.Type)
		}
		
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for action to complete")
		}
		
		time.Sleep(2 * time.Second)
	}
}