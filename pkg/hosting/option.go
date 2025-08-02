package hosting

import (
	"os"
	"time"
)

type LaunchOption func(server Server) error

func WithFileUpload(path, dest string) LaunchOption {
	return func(server Server) error {
		time.Sleep(5 * time.Second)
		_, _, err := server.Copy(path, dest)
		return err
	}
}

func WithBinaryData(path string, data []byte) LaunchOption {
	return func(server Server) error {
		time.Sleep(5 * time.Second)
		_, _, err := server.Dump(path, data)
		return err
	}
}

func WithSetupScript(script string, args ...string) LaunchOption {
	return func(server Server) error {
		time.Sleep(5 * time.Second)
		args = append([]string{script}, args...)
		return server.Connect(os.Stdin, os.Stdout, os.Stderr, args...)
	}
}
