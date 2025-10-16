package hosting

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Server represents a cloud server instance.
// A Server is created and managed by a Platform.
//
// The Platform field is set automatically when the server is created
// via Platform.NewServer or retrieved via Platform.GetServer.
// This allows server methods to delegate operations back to the platform.
type Server struct {
	// Platform is the platform that created or manages this server.
	// This field is set automatically by Platform methods.
	Platform Platform

	// ID is the unique identifier for this server within the platform.
	ID string

	// IP is the public IP address of the server.
	IP string

	// PrivIP is the private IP address of the server.
	PrivIP string

	// Loc is the geographic location or region where the server is deployed.
	// Format is platform-specific (e.g., "nyc3", "us-east-1").
	Loc string

	// Size is the server size or instance type.
	// Format is platform-specific (e.g., "s-2vcpu-4gb", "t2.micro").
	Size string

	// Name is the human-readable name of the server.
	Name string

	// Status is the current status of the server.
	// Common values include "new", "active", "off", "archive".
	Status string

	// Tags is a list of tags associated with the server.
	Tags []string
}

// Connect establishes an SSH connection to the server and executes a command.
// It uses the server's IP address and accepts new host keys automatically.
//
// The stdin, stdout, and stderr parameters control the input and output streams.
// Pass nil for any stream you don't need.
//
// The args are joined with spaces and executed as a single command on the server.
//
// Example:
//
//	var stdout bytes.Buffer
//	err := server.Connect(nil, &stdout, nil, "ls", "-la", "/var/log")
func (server *Server) Connect(stdin io.Reader, stdout, stderr io.Writer, args ...string) (err error) {
	host := fmt.Sprintf("root@%s", server.IP)
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=accept-new", host, strings.Join(args, " "))
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return errors.Wrap(cmd.Run(), strings.Join(args, " "))
}

// Exec executes a command on the server and returns the output.
// Unlike Connect, Exec captures and returns stdout and stderr as bytes.Buffer.
//
// Example:
//
//	stdout, stderr, err := server.Exec("docker", "ps", "-a")
//	if err != nil {
//		fmt.Printf("Error: %v\nStderr: %s\n", err, stderr.String())
//	}
//	fmt.Printf("Output: %s\n", stdout.String())
func (server *Server) Exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
	return stdout, stderr, server.Connect(nil, &stdout, &stderr, args...)
}

// Copy transfers a local file to the server using SCP.
// The src parameter is the local file path, and dst is the destination path on the server.
//
// Example:
//
//	stdout, stderr, err := server.Copy("./app.tar.gz", "/opt/app.tar.gz")
func (server *Server) Copy(src, dst string) (stdout, stderr bytes.Buffer, _ error) {
	dst = fmt.Sprintf("root@%s:%s", server.IP, dst)
	cmd := exec.Command("scp", "-o", "StrictHostKeyChecking=accept-new", src, dst)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, errors.Wrapf(cmd.Run(), "failed to copy %s to %s", src, dst)
}

// Dump writes data to a file on the server.
// This method creates a temporary local file, writes the data to it,
// copies it to the server, and optionally makes it executable.
//
// The path parameter is the destination file path on the server.
// If executable is true, the file will be made executable (chmod +x).
//
// Example:
//
//	script := []byte("#!/bin/bash\necho 'Hello, World!'")
//	_, _, err := server.Dump("/usr/local/bin/hello", script, true)
func (server *Server) Dump(path string, data []byte, executable bool) (stdout, stderr bytes.Buffer, err error) {
	file, err := os.CreateTemp("", "skyfile-*")
	if err != nil {
		return stdout, stderr, errors.Wrap(err, "failed to create temp file")
	}

	defer os.Remove(file.Name())
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		return stdout, stderr, errors.Wrap(err, "failed to write data to file")
	}

	if stdout, stderr, err = server.Copy(file.Name(), path); err != nil {
		return stdout, stderr, errors.Wrap(err, "failed to copy file "+path)
	}

	if executable {
		_, _, err = server.Exec("chmod", "+x", path)
		if err != nil {
			return stdout, stderr, errors.Wrap(err, "failed to chmod file")
		}
	}

	return stdout, stderr, nil
}

// Destroy permanently deletes this server.
// This is a convenience method that delegates to Platform.DestroyServer.
//
// This operation cannot be undone. The server and all its data will be lost.
//
// Example:
//
//	err := server.Destroy()
func (server *Server) Destroy() error {
	return server.Platform.DestroyServer(server.ID)
}
