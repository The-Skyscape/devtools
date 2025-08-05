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
	sshArgs := append([]string{"-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=30", "root@" + s.ip}, args...)
	// Debug logging
	fmt.Printf("🔌 SSH: %s\n", strings.Join(args, " "))
	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("❌ SSH Error: %v\n", err)
	}
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
		redeploy = flag.Bool("redeploy", false, "Redeploy to existing server")
		destroy  = flag.Bool("destroy", false, "Destroy server and remove config")
		list     = flag.Bool("list", false, "List all servers")
	)

	flag.Usage = func() {
		fmt.Print(LAUNCH_USAGE)
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle list command
	if *list {
		return listServers()
	}

	// Check for API key
	apiKey := digitalocean.ApiKey
	if apiKey == "" {
		return errors.New("DIGITAL_OCEAN_API_KEY environment variable is required")
	}

	// Handle destroy command
	if *destroy {
		return destroyServer(*name, apiKey)
	}

	// Check if binary path is provided
	if *binary == "" {
		return errors.New("--binary flag is required to specify the application binary")
	}

	var deployedServer hosting.Server
	var config *ServerConfig

	// Handle redeployment to existing server
	if *redeploy {
		// Load existing server config
		existingConfig, err := loadServerConfig(*name)
		if err != nil {
			return errors.Wrap(err, "server not found - use 'servers/' to see available servers")
		}
		
		fmt.Printf("🔄 Redeploying to existing server '%s' at %s...\n", *name, existingConfig.IP)
		
		// Create a simple server wrapper with the stored IP address
		deployedServer = &ipServer{ip: existingConfig.IP}
		config = existingConfig
	} else {
		// Check if server already exists
		if _, err := os.Open(filepath.Join("servers", *name+".json")); err == nil {
			existingConfig, err := loadServerConfig(*name)
			if err != nil {
				return errors.Wrap(err, "failed to load existing server")
			}
			fmt.Printf("Server '%s' already exists at: http://%s\n", *name, existingConfig.IP)
			fmt.Printf("To redeploy, use: --redeploy\n")
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
	if *redeploy {
		fmt.Printf("📤 Uploading application files...\n")
		
		// Upload binary and Dockerfile using server methods
		if _, _, err := deployedServer.Copy(*binary, "/root/app"); err != nil {
			return errors.Wrap(err, "failed to upload binary")
		}
		
		// Longer delay to avoid connection throttling
		time.Sleep(3 * time.Second)
		
		if _, _, err := deployedServer.Dump("/root/Dockerfile", dockerfile); err != nil {
			return errors.Wrap(err, "failed to upload Dockerfile")
		}
		
		// Longer delay after file operations
		time.Sleep(3 * time.Second)
		
		// Stop and remove existing container in a single command to reduce SSH connections
		fmt.Printf("🛑 Stopping existing container...\n")
		deployedServer.Exec("bash", "-c", "docker stop sky-app 2>/dev/null; docker rm sky-app 2>/dev/null || true")
		
		// Longer delay before building to let SSH recover
		time.Sleep(5 * time.Second)
	}


	// Build and deploy container
	fmt.Printf("🐳 Building Docker image...\n")
	host := containers.Remote(deployedServer)
	if err := host.BuildImage("skyscape:latest", "/root"); err != nil {
		return errors.Wrap(err, "failed to build Docker image")
	}
	
	// Longer delay after build to let SSH recover
	time.Sleep(5 * time.Second)

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
			"/root/app": "/app",
		},
		Env: map[string]string{
			"PORT":  "80",
			"THEME": "corporate",
		},
	}

	fmt.Printf("🚀 Launching application container...\n")
	
	// Debug: Test SSH connection before launch
	fmt.Printf("🔍 Testing SSH connection...\n")
	if _, _, err := deployedServer.Exec("echo", "SSH connection test"); err != nil {
		fmt.Printf("❌ SSH test failed: %v\n", err)
		fmt.Printf("⏳ Waiting 10 seconds for SSH to recover...\n")
		time.Sleep(10 * time.Second)
		
		// Try again
		if _, _, err := deployedServer.Exec("echo", "SSH connection test 2"); err != nil {
			fmt.Printf("❌ SSH still failing after wait: %v\n", err)
		} else {
			fmt.Printf("✅ SSH recovered after wait\n")
		}
	} else {
		fmt.Printf("✅ SSH connection OK\n")
	}
	
	if err := host.Launch(service); err != nil {
		return errors.Wrap(err, "failed to launch container")
	}
	
	// Small delay after container launch
	time.Sleep(1 * time.Second)

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
				fmt.Printf("✅ SSL certificates generated and container restarted!\n")
			}

			config.Domain = *domain
			saveServerConfig(config)
		}
	}
	
	// Save updated config for redeployment
	if *redeploy {
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
	if *redeploy {
		fmt.Printf("\n✅ Application successfully redeployed to '%s'\n", config.Name)
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

func listServers() error {
	files, err := filepath.Glob("servers/*.json")
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No servers found.")
		return nil
	}

	fmt.Println("\n📋 Configured Servers:")
	fmt.Println("────────────────────────────────────────────")
	
	for _, file := range files {
		config, err := loadServerConfig(strings.TrimSuffix(filepath.Base(file), ".json"))
		if err != nil {
			continue
		}
		
		fmt.Printf("  %-20s %s", config.Name, config.IP)
		if config.Domain != "" {
			fmt.Printf(" (%s)", config.Domain)
		}
		fmt.Printf("\n")
	}
	
	fmt.Println("────────────────────────────────────────────")
	fmt.Printf("\nTotal: %d server(s)\n", len(files))
	return nil
}

func destroyServer(name string, apiKey string) error {
	if name == "" {
		return errors.New("--name flag is required for destroy operation")
	}

	// Load server config
	config, err := loadServerConfig(name)
	if err != nil {
		return errors.Wrap(err, "server not found")
	}

	// Confirm destruction
	fmt.Printf("\n⚠️  WARNING: This will destroy the server '%s' at %s\n", config.Name, config.IP)
	if config.Domain != "" {
		fmt.Printf("   Domain: %s\n", config.Domain)
	}
	fmt.Printf("\nThis action cannot be undone. The server and all its data will be permanently deleted.\n")
	fmt.Printf("\nType the server name '%s' to confirm destruction: ", config.Name)
	
	var confirmation string
	fmt.Scanln(&confirmation)
	
	if confirmation != config.Name {
		fmt.Println("❌ Destruction cancelled.")
		return nil
	}

	fmt.Printf("\n🗑️  Destroying server '%s'...\n", config.Name)

	// Initialize platform and get server by ID
	platform := digitalocean.Connect(apiKey)
	ctx := context.Background()
	
	if config.ID != "" {
		server, err := platform.GetServer(config.ID)
		if err != nil {
			fmt.Printf("⚠️  Failed to get server from DigitalOcean: %v\n", err)
			fmt.Printf("    Server may have been deleted manually\n")
		} else {
			// Destroy the server
			if err := server.Destroy(ctx); err != nil {
				return errors.Wrap(err, "failed to destroy server")
			}
			fmt.Printf("✅ Server destroyed in DigitalOcean\n")
		}
	} else {
		fmt.Printf("⚠️  No server ID found in config (old config format)\n")
	}

	// Remove config file
	configPath := filepath.Join("servers", config.Name+".json")
	if err := os.Remove(configPath); err != nil {
		fmt.Printf("⚠️  Failed to remove config file: %v\n", err)
	} else {
		fmt.Printf("✅ Server config removed\n")
	}

	fmt.Printf("\n✅ Server '%s' has been destroyed.\n", config.Name)
	return nil
}

func main() {
	if err := runLaunch(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
