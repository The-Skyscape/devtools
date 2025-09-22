package hosting

import (
	"bytes"
	"context"
	"io"
)

type Platform interface {
	Launch(s *Server) (*Server, error)
	Server(id string) (*Server, error)
}

type Server interface {
	GetID() string
	GetIP() string
	GetName() string

	Launch(opts ...LaunchOption) error
	Destroy(ctx context.Context) error
	Alias(sub, domain string) error

	Env(string, string) error
	Exec(args ...string) (bytes.Buffer, bytes.Buffer, error)
	Copy(string, string) (bytes.Buffer, bytes.Buffer, error)
	Dump(string, []byte) (bytes.Buffer, bytes.Buffer, error)
	Connect(io.Reader, io.Writer, io.Writer, ...string) error
}
