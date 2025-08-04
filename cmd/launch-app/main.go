package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
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

// ipServer is a simple Server implementation that only needs an IP address for redeployment
type ipServer struct {
	ip string
}

func (s *ipServer) GetID() string   { return "" }
func (s *ipServer) GetIP() string   { return s.ip }
func (s *ipServer) GetName() string { return "" }

func (s *ipServer) Launch(opts ...hosting.LaunchOption) error {
	return errors.New("cannot launch an IP-only server")
}

func (s *ipServer) Destroy(ctx context.Context) error {
	return errors.New("cannot destroy an IP-only server")
}

func (s *ipServer) Alias(sub, domain string) error {
	return errors.New("cannot configure DNS for IP-only server")
}

func (s *ipServer) Env(key, value string) error {
	return nil // No-op for IP-only server
}

func (s *ipServer) Exec(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	// Execute command via SSH
	var stdout, stderr bytes.Buffer
	sshArgs := append([]string{"-o", "StrictHostKeyChecking=no", "root@" + s.ip}, args...)
	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout, stderr, err
}

func (s *ipServer) Copy(src, dst string) (bytes.Buffer, bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("scp", "-o", "StrictHostKeyChecking=no", src, "root@"+s.ip+":"+dst)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout, stderr, err
}

func (s *ipServer) Dump(path string, data []byte) (bytes.Buffer, bytes.Buffer, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "dump-*")
	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	tmpFile.Close()

	return s.Copy(tmpFile.Name(), path)
}

func (s *ipServer) Connect(stdin io.Reader, stdout io.Writer, stderr io.Writer, args ...string) error {
	sshArgs := append([]string{"-o", "StrictHostKeyChecking=no", "root@" + s.ip}, args...)
	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
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
		redeploy = flag.String("redeploy", "", "Redeploy to existing server by IP address")
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

	// Check if binary path is provided
	if *binary == "" {
		return errors.New("--binary flag is required to specify the application binary")
	}

	var deployedServer hosting.Server
	var config *ServerConfig

	// Handle redeployment to existing server by IP
	if *redeploy != "" {
		fmt.Printf("🔄 Redeploying to existing server at %s...\n", *redeploy)
		
		// Create a simple server wrapper with just the IP address
		deployedServer = &ipServer{ip: *redeploy}
		
		// Load or create config
		if existingConfig, err := loadServerConfig(*name); err == nil {
			config = existingConfig
			config.IP = *redeploy // Update IP in case it changed
		} else {
			config = &ServerConfig{
				IP:        *redeploy,
				Name:      *name,
				Provider:  *provider,
				Domain:    *domain,
				Binary:    *binary,
				CreatedAt: time.Now(),
			}
		}
	} else {
		// Check if server already exists
		if _, err := os.Open(filepath.Join("servers", *name+".json")); err == nil {
			existingConfig, err := loadServerConfig(*name)
			if err != nil {
				return errors.Wrap(err, "failed to load existing server")
			}
			fmt.Printf("Server already launched: http://%s\n", existingConfig.IP)
			fmt.Printf("To redeploy, use: --redeploy %s\n", existingConfig.IP)
			return nil
		}

		// Connect to DigitalOcean and launch new server
		fmt.Printf("☁️  Creating DigitalOcean droplet...\n")
		var err error
		deployedServer, err = digitalocean.Connect(apiKey).Launch(
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

		// Save server config for new server
		config = &ServerConfig{
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
	}

	// For redeployment, upload new binary and Dockerfile
	if *redeploy != "" {
		fmt.Printf("📤 Uploading application files...\n")
		
		// Upload binary and Dockerfile using server methods
		if _, _, err := deployedServer.Copy(*binary, "/root/app"); err != nil {
			return errors.Wrap(err, "failed to upload binary")
		}
		if _, _, err := deployedServer.Dump("/root/Dockerfile", dockerfile); err != nil {
			return errors.Wrap(err, "failed to upload Dockerfile")
		}
		
		// Stop existing container if running
		fmt.Printf("🛑 Stopping existing container...\n")
		host := containers.Remote(deployedServer)
		host.Exec("docker", "stop", "sky-app")
		host.Exec("docker", "rm", "sky-app")
	}

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
	
	// Save updated config for redeployment
	if *redeploy != "" {
		config.Binary = *binary
		if *domain != "" {
			config.Domain = *domain
		}
		saveServerConfig(config)
	}

	// Final output
	fmt.Printf("\n🎉 Deployment complete!\n\n")
	fmt.Printf("Your application is now running at:\n")
	fmt.Printf("  🔗 http://%s\n", deployedServer.GetIP())
	if config.Domain != "" {
		fmt.Printf("  🔗 https://%s\n", config.Domain)
	}
	if *redeploy != "" {
		fmt.Printf("\n✅ Application successfully redeployed to %s\n", *redeploy)
	} else {
		fmt.Printf("\n📋 Server Details:\n")
		fmt.Printf("  ID: %s\n", deployedServer.GetID())
		fmt.Printf("  IP: %s\n", deployedServer.GetIP())
		fmt.Printf("  Size: %s\n", *size)
		fmt.Printf("  Region: %s\n", *region)
	}
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

func main() {
	if err := runLaunch(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
