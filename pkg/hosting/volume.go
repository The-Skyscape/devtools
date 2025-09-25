package hosting

import "github.com/pkg/errors"

// Volume represents a persistent storage volume that can be attached to servers.
// A Volume is created and managed by a Platform.
//
// The Platform field is set automatically when the volume is created
// via Platform.NewVolume or retrieved via Platform.GetVolume.
// This allows volume methods to delegate operations back to the platform.
type Volume struct {
	// Platform is the platform that created or manages this volume.
	// This field is set automatically by Platform methods.
	Platform Platform

	// ID is the unique identifier for this volume within the platform.
	ID string

	// Loc is the geographic location or region where the volume exists.
	// Volumes can typically only be mounted to servers in the same location.
	Loc string

	// Size is the capacity of the volume in gigabytes.
	Size int

	// Name is the human-readable name of the volume.
	Name string
}

// Mount attaches this volume to the specified server.
// This is a convenience method that delegates to Platform.MountVolume.
//
// The volume and server must be in the same location.
// A volume can typically only be mounted to one server at a time.
//
// Example:
//	err := volume.Mount(server)
func (volume *Volume) Mount(server *Server) (err error) {
	err = volume.Platform.MountVolume(volume, server)
	return errors.Wrap(err, "failed to mount volume")
}

// Unmount detaches this volume from the specified server.
// This is a convenience method that delegates to Platform.UnmountVolume.
//
// The volume must currently be mounted to the specified server.
//
// Example:
//	err := volume.Unmount(server)
func (volume *Volume) Unmount(server *Server) (err error) {
	err = volume.Platform.UnmountVolume(volume, server)
	return errors.Wrap(err, "failed to unmount volume")
}
