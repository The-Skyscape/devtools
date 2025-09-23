package hosting

import "github.com/pkg/errors"

type Volume struct {
	Platform Platform
	ID       string
	Loc      string
	Size     int // in bytes
	Name     string
}

func (volume *Volume) Mount(server *Server) (err error) {
	err = volume.Platform.MountVolume(volume, server)
	return errors.Wrap(err, "failed to mount volume")
}

func (volume *Volume) Unmount(server *Server) (err error) {
	err = volume.Platform.UnmountVolume(volume, server)
	return errors.Wrap(err, "failed to unmount volume")
}
