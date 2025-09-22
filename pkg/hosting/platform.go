package hosting

import (
	"bytes"
	"context"
	"io"
)

// Platform defines the minimal interface for hosting platforms
type Platform interface {
	// Server operations
	Launch(opts ...interface{}) (Server, error)
	Server(id string) (Server, error)
	Servers() ([]Server, error)

	// Volume operations
	Volume(id string) (Volume, error)
	Volumes() ([]Volume, error)

	// Domain operations
	Domain(name string) (Domain, error)
	Domains() ([]Domain, error)
}

// Server represents a compute instance
type Server interface {
	// Fields
	ID() string
	IP() string
	Name() string

	// Volumes
	Mount(Volume) (Volume, error)
	Volumes() ([]Volume, error)

	// Domains
	Alias(Domain) (Domain, error)
	Domains() ([]Domain, error)

	// Actions
	Env(key, value string) error
	Copy(string, string) (bytes.Buffer, bytes.Buffer, error)
	Dump(string, []byte) (bytes.Buffer, bytes.Buffer, error)
	Exec(args ...string) (bytes.Buffer, bytes.Buffer, error)
	Connect(io.Reader, io.Writer, io.Writer, ...string) error
	Destroy(context.Context) error
}

// Volume represents a block storage volume
type Volume interface {
	ID() string
	Name() string

	Server() (Server, error)
}

// Domain represents a DNS domain and record
type Domain interface {
	ID() string
	Type() string
	Name() string
	Data() string

	Server() (Server, error)
}