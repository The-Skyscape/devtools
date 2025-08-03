package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/containers"
	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"
	"github.com/pkg/errors"
)

type ServerConfig struct {
	ID        string
	Name      string
	IP        string
	Size      string
	Region    string
	Provider  string
	Domain    string
	Binary    string
	CreatedAt time.Time
}

const LAUNCH_USAGE = `
TheSkyscape DevTools Launch Command Usage:

  $ launch-app [options]

Options:

`

var (
	//go:embed resources/Dockerfile
	dockerfile []byte

	//go:embed resources/generate-certs.sh
	generateCerts string

	//go:embed resources/setup-server.sh
	setupServer string
)

func runLaunch() error {
	var (
		provider = flag.String("provider", "digitalocean", "Cloud provider (digitalocean)")
		region   = flag.String("region", "sfo3", "Server region")
		size     = flag.String("size", "s-2vcpu-2gb", "Server size")
		domain   = flag.String("domain", "", "Domain name for SSL (optional)")
		name     = flag.String("name", "skyscape-app", "Server name")
		binary   = flag.String("binary", "", "Path to application binary")
	)

	flag.Usage = func() {
		fmt.Print(LAUNCH_USAGE)
		flag.PrintDefaults()
	}

	flag.Parse()

	// Check for API key
	apiKey := digitalocean.ApiKey
	if apiKey == "" {
		return errors.New("DIGITAL_OCEAN_API_KEY environment variable is required")
	}

	// Check if server already exists
	if _, err := os.Open(filepath.Join("servers", *name+".json")); err == nil {
		config, err := loadServerConfig(*name)
		if err != nil {
			return errors.Wrap(err, "failed to load existing server")
		}
		fmt.Printf("Server already launched: http://%s\n", config.IP)
		return nil
	}

	// Connect to DigitalOcean and launch server
	fmt.Printf("☁️  Creating DigitalOcean droplet...\n")
	deployedServer, err := digitalocean.Connect(apiKey).Launch(
		&digitalocean.Server{
			Name:   *name,
			Size:   *size,
			Region: *region,
			Image:  "docker-20-04",
			Status: "new",
		},
		hosting.WithBinaryData("/root/Dockerfile", dockerfile),
		hosting.WithFileUpload(*binary, "/root/app"),
		hosting.WithSetupScript(setupServer),
	)

	if err != nil {
		return errors.Wrap(err, "failed to launch server")
	}

	// Save server config
	config := &ServerConfig{
		ID:        deployedServer.GetID(),
		IP:        deployedServer.GetIP(),
		Name:      *name,
		Size:      *size,
		Region:    *region,
		Provider:  *provider,
		Domain:    *domain,
		Binary:    *binary,
		CreatedAt: time.Now(),
	}

	if err := saveServerConfig(config); err != nil {
		return errors.Wrap(err, "failed to save server config")
	}

	fmt.Printf("✅ Server created successfully!\n")
	fmt.Printf("📍 Server ID: %s\n", deployedServer.GetID())
	fmt.Printf("🌍 IP Address: %s\n", deployedServer.GetIP())

	// Build and deploy container
	fmt.Printf("🐳 Building Docker image...\n")
	host := containers.Remote(deployedServer)
	if err := host.BuildImage("skyscape:latest", "."); err != nil {
		return errors.Wrap(err, "failed to build Docker image")
	}

	// Create and launch service
	service := &containers.Service{
		Privileged: true,
		Name:       "sky-app",
		Image:      "skyscape:latest",
		Entrypoint: "/app",
		Network:    "host",
		Mounts: map[string]string{
			"/root/.skyscape":      "/root/.skyscape",
			"/var/run/docker.sock": "/var/run/docker.sock",
		},
		Copied: map[string]string{
			"/root/app":           "/app",
			"/root/fullchain.pem": "/root/fullchain.pem",
			"/root/privkey.pem":   "/root/privkey.pem",
		},
		Env: map[string]string{
			"PORT":  "80",
			"THEME": "corporate",
		},
	}

	fmt.Printf("🚀 Launching application container...\n")
	if err := host.Launch(service); err != nil {
		return errors.Wrap(err, "failed to launch container")
	}

	// Configure domain if provided
	if parts := strings.SplitN(*domain, ".", 2); len(parts) == 2 {
		sub, root := parts[0], parts[1]
		fmt.Printf("🌐 Configuring domain: %s.%s\n", sub, root)
		if err := deployedServer.Alias(sub, root); err != nil {
			fmt.Printf("⚠️  Domain configuration failed: %v\n", err)
		} else {
			fmt.Printf("✅ Domain configured successfully!\n")

			// Generate SSL certificates
			fmt.Printf("🔒 Generating SSL certificates...\n")
			certScript := fmt.Sprintf(generateCerts, *domain, "admin@"+*domain, apiKey)
			if _, _, err := deployedServer.Exec(certScript); err != nil {
				fmt.Printf("⚠️  SSL certificate generation failed: %v\n", err)
			} else {
				fmt.Printf("✅ SSL certificates generated!\n")
				if err = service.Stop(); err != nil {
					fmt.Printf("⚠️  Server stop failed: %v\n", err)
				} else if err = service.Start(); err != nil {
					fmt.Printf("⚠️  Server restart failed: %v\n", err)
				}
			}

			config.Domain = *domain
			saveServerConfig(config)
		}
	}

	// Final output
	fmt.Printf("\n🎉 Deployment complete!\n\n")
	fmt.Printf("Your application is now running at:\n")
	fmt.Printf("  🔗 http://%s\n", deployedServer.GetIP())
	if *domain != "" {
		fmt.Printf("  🔗 https://%s\n", *domain)
	}
	fmt.Printf("\n📋 Server Details:\n")
	fmt.Printf("  ID: %s\n", deployedServer.GetID())
	fmt.Printf("  IP: %s\n", deployedServer.GetIP())
	fmt.Printf("  Size: %s\n", *size)
	fmt.Printf("  Region: %s\n", *region)
	fmt.Printf("\n📝 To connect via SSH:\n")
	fmt.Printf("  ssh root@%s\n", deployedServer.GetIP())

	return nil
}

func loadServerConfig(serverName string) (*ServerConfig, error) {
	configPath := filepath.Join("servers", serverName+".json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveServerConfig(config *ServerConfig) error {
	if err := os.MkdirAll("servers", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join("servers", config.Name+".json")
	return os.WriteFile(configPath, data, 0644)
}

func init() {
	// Initialize any required setup here if needed
}

func main() {
	if err := runLaunch(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
