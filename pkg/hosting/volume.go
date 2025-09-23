package hosting

import "github.com/pkg/errors"

type Volume struct {
	platform Platform
	Loc      string
	Size     int // in bytes
	Name     string
	ServerID string
}

func (volume *Volume) Mount(server *Server) (err error) {
	err = volume.platform.MountVolume(volume, server)
	return errors.Wrap(err, "failed to mount volume")
}
