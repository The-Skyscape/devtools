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

type Server struct {
	platform Platform
	ID       string
	IP       string
	Loc      string
	Size     string
	Name     string
	Status   string
}

func (server *Server) Connect(stdin io.Reader, stdout, stderr io.Writer, args ...string) (err error) {
	host := fmt.Sprintf("root@%s", server.IP)
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=accept-new", host, strings.Join(args, " "))
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return errors.Wrap(cmd.Run(), strings.Join(args, " "))
}

func (server *Server) Env(key, value string) error {
	return server.Connect(nil, nil, nil, "echo \"export $key=$value\" >> ~/.bashrc")
}

func (server *Server) Exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
	return stdout, stderr, server.Connect(nil, &stdout, &stderr, args...)
}

func (server *Server) Copy(src, dst string) (stdout, stderr bytes.Buffer, _ error) {
	dst = fmt.Sprintf("root@%s:%s", server.IP, dst)
	cmd := exec.Command("scp", "-o", "StrictHostKeyChecking=accept-new", src, dst)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, errors.Wrapf(cmd.Run(), "failed to copy %s to %s", src, dst)
}

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

func (server *Server) Destroy() error {
	return server.platform.DestroyServer(server.ID)
}
