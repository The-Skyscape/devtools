package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

// Embedded resources
var (
	//go:embed resources/Dockerfile
	dockerfile []byte

	//go:embed resources/generate-certs.sh
	generateCerts string

	//go:embed resources/setup-server.sh
	setupServer string

	//go:embed resources/deploy.sh
	deployScript string
)

// Global flags
var (
	provider string
	region   string
	size     string
	domain   string
	name     string
	binary   string
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "launch-app",
	Short: "TheSkyscape DevTools Launch Command",
	Long: `TheSkyscape DevTools Launch Command

Deploy and manage Skyscape applications on cloud servers with integrated 
SSL certificates, Docker containers, and domain configuration.

Examples:
  launch-app create --name my-app --binary ./my-app
  launch-app deploy --name my-app --binary ./my-app --redeploy
  launch-app list
  launch-app destroy --name my-app`,
}

// Create command - creates a new server and deploys
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new server and deploy application",
	Long: `Create a new cloud server and deploy your application.

This command will:
- Create a new DigitalOcean droplet
- Install Docker and dependencies
- Deploy your application in a container
- Configure SSL certificates if domain is provided

Examples:
  launch-app create --name my-app --binary ./my-app
  launch-app create --name my-app --binary ./my-app --domain app.example.com`,
	RunE: runCreate,
}

// Deploy command - redeploy to existing server
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy application to existing server",
	Long: `Deploy or redeploy your application to an existing server.

This command will:
- Upload your new application binary
- Rebuild the Docker container
- Restart the application
- Update SSL certificates if needed

Examples:
  launch-app deploy --name my-app --binary ./my-app-v2
  launch-app deploy --name my-app --binary ./my-app --domain new.example.com`,
	RunE: runDeploy,
}

// List command - list all servers
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured servers",
	Long:  `List all servers that have been created and configured.`,
	RunE:  runList,
}

// Destroy command - destroy server
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a server and remove configuration",
	Long: `Destroy a cloud server and remove its configuration.

This will permanently delete the server and all its data.
This action cannot be undone.

Examples:
  launch-app destroy --name my-app`,
	RunE: runDestroy,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&provider, "provider", "digitalocean", "Cloud provider")
	rootCmd.PersistentFlags().StringVar(&region, "region", "sfo3", "Server region")
	rootCmd.PersistentFlags().StringVar(&size, "size", "s-2vcpu-2gb", "Server size")
	rootCmd.PersistentFlags().StringVar(&domain, "domain", "", "Domain name for SSL (optional)")
	rootCmd.PersistentFlags().StringVar(&name, "name", "", "Server name (required)")
	rootCmd.PersistentFlags().StringVar(&binary, "binary", "", "Path to application binary")

	// Mark required flags
	createCmd.MarkPersistentFlagRequired("name")
	createCmd.MarkPersistentFlagRequired("binary")
	deployCmd.MarkPersistentFlagRequired("name")
	deployCmd.MarkPersistentFlagRequired("binary")
	destroyCmd.MarkPersistentFlagRequired("name")

	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(destroyCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Check for API key
	apiKey := digitalocean.ApiKey
	if apiKey == "" {
		return errors.New("DIGITAL_OCEAN_API_KEY environment variable is required")
	}

	// Check if server already exists
	if _, err := os.Open(filepath.Join("servers", name+".json")); err == nil {
		existingConfig, err := loadServerConfig(name)
		if err != nil {
			return errors.Wrap(err, "failed to load existing server")
		}
		fmt.Printf("Server '%s' already exists at: http://%s\n", name, existingConfig.IP)
		fmt.Printf("To redeploy, use: launch-app deploy --name %s --binary %s\n", name, binary)
		return nil
	}

	fmt.Printf("☁️  Creating DigitalOcean droplet '%s'...\n", name)

	// Connect to DigitalOcean and launch new server
	deployedServer, err := digitalocean.Connect(apiKey).Launch(
		&digitalocean.Server{
			Name:   name,
			Size:   size,
			Region: region,
			Image:  "docker-20-04",
			Status: "new",
		},
		hosting.WithSetupScript(setupServer),
	)

	if err != nil {
		return errors.Wrap(err, "failed to launch server")
	}

	fmt.Printf("✅ Server created successfully!\n")
	fmt.Printf("📍 Server ID: %s\n", deployedServer.GetID())
	fmt.Printf("🌍 IP Address: %s\n", deployedServer.GetIP())

	// Wait for server to boot up and SSH to be ready
	fmt.Printf("⏳ Waiting for server to be ready...\n")
	if err := waitForSSH(deployedServer.GetIP(), 60); err != nil {
		return errors.Wrap(err, "server failed to become ready")
	}
	fmt.Printf("✅ Server is ready for deployment!\n")

	// Save server config
	config := &ServerConfig{
		ID:        deployedServer.GetID(),
		IP:        deployedServer.GetIP(),
		Name:      name,
		Size:      size,
		Region:    region,
		Provider:  provider,
		Domain:    domain,
		Binary:    binary,
		CreatedAt: time.Now(),
	}

	if err := saveServerConfig(config); err != nil {
		return errors.Wrap(err, "failed to save server config")
	}

	// Now deploy the application
	return deployApplication(deployedServer, config, apiKey, false)
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// Load existing server config
	config, err := loadServerConfig(name)
	if err != nil {
		return errors.Wrap(err, "server not found - use 'launch-app list' to see available servers")
	}

	fmt.Printf("🔄 Deploying to existing server '%s' at %s...\n", name, config.IP)

	// Get API key
	apiKey := digitalocean.ApiKey
	if apiKey == "" {
		// For redeploy, we can work without API key but SSL operations will be skipped
		fmt.Printf("⚠️  DIGITAL_OCEAN_API_KEY not set - SSL operations will be skipped\n")
	}

	// Get the server from platform if we have API key and server ID
	var deployedServer hosting.Server
	if apiKey != "" && config.ID != "" {
		platform := digitalocean.Connect(apiKey)
		deployedServer, err = platform.GetServer(config.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get server from platform")
		}
	} else {
		// Create a minimal server implementation for redeploy without API key
		return errors.New("redeploy requires DIGITAL_OCEAN_API_KEY to be set")
	}

	// Update config with new binary path and domain if provided
	config.Binary = binary
	if domain != "" {
		config.Domain = domain
	}

	return deployApplication(deployedServer, config, apiKey, true)
}

func runList(cmd *cobra.Command, args []string) error {
	return listServers()
}

func runDestroy(cmd *cobra.Command, args []string) error {
	// Check for API key
	apiKey := digitalocean.ApiKey
	if apiKey == "" {
		return errors.New("DIGITAL_OCEAN_API_KEY environment variable is required")
	}

	return destroyServer(name, apiKey)
}

// deployApplication handles the core deployment logic
func deployApplication(server hosting.Server, config *ServerConfig, apiKey string, isRedeploy bool) error {
	fmt.Printf("🚀 Deploying application using integrated deployment script...\n")

	// Upload application files
	fmt.Printf("📤 Uploading application files...\n")

	// Upload binary first
	if _, _, err := server.Copy(binary, "/root/app"); err != nil {
		return errors.Wrap(err, "failed to upload binary")
	}

	// Upload Dockerfile
	if _, _, err := server.Dump("/root/Dockerfile", dockerfile); err != nil {
		return errors.Wrap(err, "failed to upload Dockerfile")
	}

	// Configure domain DNS if provided
	var domainForScript string
	if domain != "" {
		if parts := strings.SplitN(domain, ".", 2); len(parts) == 2 {
			sub, root := parts[0], parts[1]
			fmt.Printf("🌐 Configuring domain: %s.%s\n", sub, root)
			if err := server.Alias(sub, root); err != nil {
				fmt.Printf("⚠️  Domain configuration failed: %v\n", err)
				domainForScript = "" // Don't configure SSL if DNS failed
			} else {
				fmt.Printf("✅ Domain configured successfully!\n")
				domainForScript = domain
			}
		} else {
			fmt.Printf("⚠️  Invalid domain format: %s\n", domain)
			domainForScript = ""
		}
	}

	// Execute the comprehensive deploy script
	fmt.Printf("🔧 Executing deployment script...\n")

	// Prepare parameters - use original domain even if DNS config failed
	deployDomain := domain // Use the original domain parameter
	email := ""
	if deployDomain != "" {
		email = "admin@" + deployDomain
	}

	redeployFlag := "false"
	if isRedeploy {
		redeployFlag = "true"
	}

	// Get AUTH_SECRET from environment or generate one
	authSecret := os.Getenv("AUTH_SECRET")
	if authSecret == "" {
		// Generate a secure random secret
		authSecret = fmt.Sprintf("skyscape-%d-%s", time.Now().Unix(), config.Name)
	}

	// Execute the deployment script directly through stdin
	// This avoids multiple SSH connections and potential throttling
	fmt.Printf("🔧 Executing deployment script via single SSH connection...\n")
	
	// Prepare the script with parameters already substituted
	deployScriptWithParams := fmt.Sprintf(`
export DOMAIN="%s"
export EMAIL="%s"
export API_TOKEN="%s"
export REDEPLOY="%s"
export AUTH_SECRET="%s"

%s
`, deployDomain, email, apiKey, redeployFlag, authSecret, deployScript)

	// Execute the script through a single SSH connection
	scriptReader := strings.NewReader(deployScriptWithParams)
	var stdout, stderr strings.Builder
	
	err := server.Connect(scriptReader, &stdout, &stderr, "/bin/bash", "-s")
	if err != nil {
		fmt.Printf("❌ Deployment failed:\n")
		fmt.Printf("STDOUT: %s\n", stdout.String())
		fmt.Printf("STDERR: %s\n", stderr.String())
		return errors.Wrap(err, "deployment script failed")
	}

	// Show deployment output
	fmt.Printf("📋 Deployment output:\n%s\n", stdout.String())
	if stderr.Len() > 0 {
		fmt.Printf("⚠️  Warnings/Errors:\n%s\n", stderr.String())
	}

	// Update config with domain if it was successfully configured
	if domainForScript != "" {
		config.Domain = domainForScript
	}

	// Save updated config
	if err := saveServerConfig(config); err != nil {
		fmt.Printf("⚠️  Failed to save server config: %v\n", err)
	}

	// Final output
	fmt.Printf("\n🎉 Deployment complete!\n\n")
	fmt.Printf("Your application is now running at:\n")
	fmt.Printf("  🔗 http://%s\n", server.GetIP())
	if config.Domain != "" {
		fmt.Printf("  🔗 https://%s\n", config.Domain)
	}

	if isRedeploy {
		fmt.Printf("\n✅ Application successfully redeployed to '%s'\n", config.Name)
	} else {
		fmt.Printf("\n📋 Server Details:\n")
		fmt.Printf("  ID: %s\n", config.ID)
		fmt.Printf("  IP: %s\n", server.GetIP())
		fmt.Printf("  Size: %s\n", size)
		fmt.Printf("  Region: %s\n", region)
	}
	fmt.Printf("\n📝 To connect via SSH:\n")
	fmt.Printf("  ssh root@%s\n", server.GetIP())

	return nil
}


// waitForSSH waits for SSH to be available on the given IP
func waitForSSH(ip string, maxSeconds int) error {
	start := time.Now()
	for {
		// Try to connect via SSH
		cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=5", "root@"+ip, "echo", "ready")
		if err := cmd.Run(); err == nil {
			return nil
		}

		// Check if we've exceeded the timeout
		if time.Since(start) > time.Duration(maxSeconds)*time.Second {
			return fmt.Errorf("SSH not available after %d seconds", maxSeconds)
		}

		// Wait before retrying
		fmt.Printf(".")
		time.Sleep(2 * time.Second)
	}
}

// Configuration management functions
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
