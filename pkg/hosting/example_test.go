package hosting_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"
	"github.com/The-Skyscape/devtools/pkg/hosting/platforms/mock"
)

func ExamplePlatform_Launch() {
	// Create a mock platform for testing
	platform := mock.New()

	// Launch a server with mock-specific options
	server, err := platform.Launch(
		mock.WithName("my-app-server"),
		mock.WithIP("192.168.1.100"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server launched: %s at %s\n", server.Name(), server.IP())
}

func ExamplePlatform_realWorld() {
	// Real-world usage with DigitalOcean
	apiKey := os.Getenv("DIGITAL_OCEAN_API_KEY")
	if apiKey == "" {
		log.Fatal("DIGITAL_OCEAN_API_KEY not set")
	}

	// Create a DigitalOcean platform
	platform := digitalocean.NewPlatform(apiKey)

	// Launch a server with DigitalOcean-specific options
	server, err := platform.Launch(
		digitalocean.WithName("production-app"),
		digitalocean.WithSize("s-4vcpu-8gb"),
		digitalocean.WithRegion("nyc3"),
		digitalocean.WithProject("my-project-id"),
		digitalocean.WithTags("production", "web"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Deploy application
	_, _, err = server.Copy("/local/app", "/root/app")
	if err != nil {
		log.Fatal(err)
	}

	// Run setup commands
	stdout, stderr, err := server.Exec("chmod", "+x", "/root/app")
	if err != nil {
		log.Printf("stderr: %s", stderr.String())
		log.Fatal(err)
	}
	fmt.Printf("Setup output: %s", stdout.String())

	// Set environment variables
	err = server.Env("PORT", "8080")
	if err != nil {
		log.Fatal(err)
	}

	// Create a domain alias
	domain := &mockDomain{name: "app.example.com"}
	_, err = server.Alias(domain)
	if err != nil {
		log.Fatal(err)
	}

	// List attached volumes
	volumes, err := server.Volumes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server has %d volumes attached\n", len(volumes))

	// Clean up (commented out to prevent accidental deletion)
	// err = server.Destroy(context.Background())
}

func ExampleServer_Mount() {
	// Create a mock platform
	platform := mock.New()

	// Launch a server
	server, err := platform.Launch()
	if err != nil {
		log.Fatal(err)
	}

	// Create and mount a volume
	volume := &mockVolume{id: "vol-123", name: "data-volume"}
	mountedVolume, err := server.Mount(volume)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Volume %s mounted to server\n", mountedVolume.Name())

	// List all volumes
	volumes, err := server.Volumes()
	if err != nil {
		log.Fatal(err)
	}

	for _, vol := range volumes {
		fmt.Printf("Volume: %s\n", vol.Name())
	}
}

func ExamplePlatform_traversal() {
	// This example shows the tree-like traversal capability
	platform := mock.New()

	// Launch multiple servers
	webServer, _ := platform.Launch(mock.WithName("web-server"))
	dbServer, _ := platform.Launch(mock.WithName("db-server"))

	// Get all servers
	servers, _ := platform.Servers()
	for _, server := range servers {
		fmt.Printf("Server: %s (%s)\n", server.Name(), server.IP())

		// Get volumes for each server
		volumes, _ := server.Volumes()
		for _, vol := range volumes {
			fmt.Printf("  Volume: %s\n", vol.Name())

			// Navigate back from volume to server
			volServer, _ := vol.Server()
			if volServer != nil {
				fmt.Printf("    Attached to: %s\n", volServer.Name())
			}
		}

		// Get domains for each server
		domains, _ := server.Domains()
		for _, dom := range domains {
			fmt.Printf("  Domain: %s -> %s\n", dom.Name(), dom.Data())

			// Navigate back from domain to server
			domServer, _ := dom.Server()
			if domServer != nil {
				fmt.Printf("    Points to: %s\n", domServer.Name())
			}
		}
	}

	// Clean up
	_ = webServer.Destroy(context.Background())
	_ = dbServer.Destroy(context.Background())
}

// Mock implementations for example purposes
type mockVolume struct {
	id   string
	name string
}

func (v *mockVolume) ID() string                      { return v.id }
func (v *mockVolume) Name() string                    { return v.name }
func (v *mockVolume) Server() (hosting.Server, error) { return nil, nil }

type mockDomain struct {
	name string
}

func (d *mockDomain) ID() string                      { return "dom-1" }
func (d *mockDomain) Type() string                    { return "A" }
func (d *mockDomain) Name() string                    { return d.name }
func (d *mockDomain) Data() string                    { return "10.0.0.1" }
func (d *mockDomain) Server() (hosting.Server, error) { return nil, nil }