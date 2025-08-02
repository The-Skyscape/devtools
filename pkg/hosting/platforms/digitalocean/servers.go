package digitalocean

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"github.com/The-Skyscape/devtools/pkg/hosting"
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

type Server struct {
	client *DigitalOceanClient
	ID     int
	Name   string
	Size   string
	Region string
	Image  string
	Status string
	IP     string
}

func (server *Server) GetID() string {
	return strconv.Itoa(server.ID)
}

func (server *Server) GetIP() string {
	return server.IP
}

func (server *Server) GetName() string {
	return server.Name
}

func (server *Server) load() error {
	ctx := context.Background()
	if server.ID == 0 {
		return errors.New("missing droplet id from server")
	} else if droplet, _, err := server.client.Droplets.Get(ctx, server.ID); err != nil {
		return errors.Wrap(err, "failed to get droplet")
	} else {
		server.Name = droplet.Name
		server.Size = droplet.SizeSlug
		server.Region = droplet.Region.Name
		server.Image = droplet.Image.Name
		server.Status = droplet.Status
		server.IP, _ = droplet.PublicIPv4()
	}
	return nil
}

func (server *Server) Launch(opts ...hosting.LaunchOption) (err error) {
	ctx := context.Background()

	if server.ID != 0 {
		return errors.New("server already launched")
	}

	var accessKey *godo.Key
	if accessKey, err = server.accessKey(); err != nil {
		return errors.Wrap(err, "failed to get access key")
	}

	droplet, _, err := server.client.Droplets.Create(ctx, &godo.DropletCreateRequest{
		Name:    server.Name,
		Region:  server.Region,
		Size:    server.Size,
		Image:   godo.DropletCreateImage{Slug: server.Image},
		SSHKeys: []godo.DropletCreateSSHKey{{Fingerprint: accessKey.Fingerprint}},
	})

	if err != nil {
		return errors.Wrap(err, "failed to create droplet")
	}

	server.ID = droplet.ID
	server.Status = droplet.Status
	for server.IP == "" {
		time.Sleep(10 * time.Second)
		if err = server.load(); err != nil {
			return errors.Wrap(err, "failed to get droplet")
		}
	}

	time.Sleep(30 * time.Second)
	for _, opt := range opts {
		if err := opt(server); err != nil {
			return errors.Wrap(err, "failed to apply option")
		}
	}

	return
}

func (server *Server) Destroy(ctx context.Context) (err error) {
	_, err = server.client.Droplets.Delete(ctx, server.ID)
	return errors.Wrap(err, "failed to delete droplet")
}

func (s *Server) Dump(path string, data []byte) (stdout, stderr bytes.Buffer, err error) {
	file, err := os.CreateTemp("", "skyfile-*")
	if err != nil {
		return stdout, stderr, errors.Wrap(err, "failed to create temp file")
	}

	defer os.Remove(file.Name())
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		return stdout, stderr, errors.Wrap(err, "failed to write data to file")
	}

	if stdout, stderr, err = s.Copy(file.Name(), path); err != nil {
		return stdout, stderr, errors.Wrap(err, "failed to copy file "+path)
	}

	_, _, err = s.Exec("chmod", "+x", path)
	return stdout, stderr, errors.Wrap(err, "failed to chmod file")
}

func (server *Server) Copy(path, dst string) (stdout, stderr bytes.Buffer, _ error) {
	dst = fmt.Sprintf("root@%s:%s", server.IP, dst)
	cmd := exec.Command("scp", "-o", "StrictHostKeyChecking=no", path, dst)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, errors.Wrapf(cmd.Run(), "failed to copy %s to %s", path, dst)
}

func (server *Server) Env(key, value string) error {
	_, _, err := server.Exec("echo \"export $key=$value\" >> ~/.bashrc")
	return errors.Wrap(err, "failed to export key")
}

func (server *Server) Exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
	return stdout, stderr, server.Connect(nil, &stdout, &stderr, args...)
}

func (server *Server) Connect(stdin io.Reader, stdout, _ io.Writer, args ...string) (err error) {
	var stderr bytes.Buffer
	host := fmt.Sprintf("root@%s", server.IP)
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", host, strings.Join(args, " "))
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	return errors.Wrap(cmd.Run(), strings.Join(args, " ")+" "+stderr.String())
}
