package skyapp

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	root "sky-castle"
	"sky-castle/cmd/skyapp/resources"
	"sky-castle/internal/hosting"
	"sky-castle/internal/hosting/platforms/digitalocean"
	"sky-castle/internal/services"

	"github.com/pkg/errors"
)

const LAUNCH_USAGE = `
SkyCastle App Launch Command Usage:

  $ skyapp launch [options]

Options:

`

func RunLaunch(args []string) (err error) {
	var (
		cmd     = flag.NewFlagSet("launch", flag.ExitOnError)
		name    = cmd.String("name", "host", "Name of the server")
		size    = cmd.String("size", "s-2vcpu-2gb", "Size of the server")
		region  = cmd.String("region", "sfo2", "Region of the server")
		image   = cmd.String("image", "docker-20-04", "Image of the server")
		website = cmd.String("website", "", "Name of the website to launch")

		path, build string
	)

	cmd.Usage = func() {
		fmt.Print(LAUNCH_USAGE)
		cmd.PrintDefaults()
	}

	if err = cmd.Parse(args[1:]); err != nil {
		cmd.Usage()
		return
	}

	app := &Application{Website: *website}
	if path, build, err = app.loadPath(); err != nil {
		return
	}

	if _, err = os.Open(filepath.Join(path, *name+".json")); err == nil {
		var s *ServerWithDomain
		if s, err = loadServerFromConfig(path, *name); err != nil {
			return errors.Wrap(err, "failed to load existing server")
		} else {
			log.Printf("Server already launched: http://%s", s.GetIP())
			return nil
		}
	}

	client := digitalocean.Connect("")

	log.Printf("Launching server...")

	s := &ServerWithDomain{}
	s.Server, err = client.Launch(
		&digitalocean.Server{
			Name:   *name,
			Size:   *size,
			Region: *region,
			Image:  *image,
			Status: "new",
		},
		hosting.WithBinaryData("/root/Dockerfile", root.Dockerfile),
		hosting.WithFileUpload(build, "/root/app"),
		hosting.WithSetupScript(resources.SetupServer),
	)

	if err != nil {
		return errors.Wrap(err, "failed to launch hosting server")
	}

	if err = saveConfigToFile(s, path, *name); err != nil {
		return errors.Wrap(err, "failed to save initial config")
	}

	host := services.Remote(s.Server)
	if err = host.BuildImage("sky-castle:latest", "."); err != nil {
		return errors.Wrap(err, "failed to build docker image")
	}

	service := &services.Service{
		Privileged: true,
		Name:       "sky-app",
		Image:      "sky-castle:latest",
		Entrypoint: "/app",
		Network:    "host",
		Mounts: map[string]string{
			"/root/.sky-castle":    "/root/.sky-castle",
			"/var/run/docker.sock": "/var/run/docker.sock",
		},
		Copied: map[string]string{
			"/root/app":           "/app",
			"/root/fullchain.pem": "/root/fullchain.pem",
			"/root/privkey.pem":   "/root/privkey.pem",
		},
		Env: map[string]string{
			"PORT":  "80",
			"THEME": "pastel",
		},
	}

	if err = host.Launch(service); err != nil {
		return errors.Wrap(err, "failed to launch container")
	}

	log.Printf("Server launched: http://%s", s.GetIP())
	return saveConfigToFile(s, path, *name)
}
