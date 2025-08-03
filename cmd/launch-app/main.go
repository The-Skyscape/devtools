package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"
)

var (
	provider     string
	region       string
	size         string
	domain       string
	buildCmd     string
	setupScript  string
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "launch-app [project-directory]",
	Short: "Deploy your TheSkyscape DevTools application to the cloud",
	Long: `Deploy your TheSkyscape DevTools application to cloud providers like DigitalOcean, AWS, or GCP.

This tool builds your application, creates a cloud server, uploads your app, 
and configures everything needed for production deployment.

Supported providers:
  digitalocean - DigitalOcean Droplets (default)
  aws         - AWS EC2 (coming soon)
  gcp         - Google Cloud Compute (coming soon)

Examples:
  launch-app .                                    # Deploy current directory to DigitalOcean
  launch-app my-project --domain=myapp.com        # Deploy with custom domain
  launch-app . --provider=digitalocean --size=s-2vcpu-4gb --region=nyc1`,
	Args: cobra.MaximumNArgs(1),
	Run:  launchApp,
}

func init() {
	rootCmd.Flags().StringVarP(&provider, "provider", "p", "digitalocean", "Cloud provider (digitalocean, aws, gcp)")
	rootCmd.Flags().StringVarP(&region, "region", "r", "nyc1", "Deployment region")
	rootCmd.Flags().StringVarP(&size, "size", "s", "s-2vcpu-4gb", "Server size")
	rootCmd.Flags().StringVarP(&domain, "domain", "d", "", "Custom domain to configure")
	rootCmd.Flags().StringVar(&buildCmd, "build-cmd", "go build -o app .", "Command to build the application")
	rootCmd.Flags().StringVar(&setupScript, "setup-script", "", "Custom setup script to run on server")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func launchApp(cmd *cobra.Command, args []string) {
	// Determine project directory
	projectDir := "."
	if len(args) > 0 {
		projectDir = args[0]
	}
	
	absPath, err := filepath.Abs(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving project path: %v\n", err)
		os.Exit(1)
	}
	
	projectName := filepath.Base(absPath)
	
	logStep("🚀 Deploying %s to %s", projectName, provider)
	
	// Validate project directory
	if err := validateProject(absPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Build application
	logStep("📦 Building application...")
	if err := buildApplication(absPath); err != nil {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
		os.Exit(1)
	}
	
	// Deploy based on provider
	switch provider {
	case "digitalocean":
		if err := deployToDigitalOcean(absPath, projectName); err != nil {
			fmt.Fprintf(os.Stderr, "Deployment failed: %v\n", err)
			os.Exit(1)
		}
	case "aws":
		fmt.Fprintf(os.Stderr, "AWS deployment coming soon!\n")
		os.Exit(1)
	case "gcp":
		fmt.Fprintf(os.Stderr, "GCP deployment coming soon!\n")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", provider)
		os.Exit(1)
	}
}

func validateProject(projectDir string) error {
	// Check if it's a Go project
	if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); os.IsNotExist(err) {
		return fmt.Errorf("not a Go project (no go.mod found)")
	}
	
	// Check for main.go
	if _, err := os.Stat(filepath.Join(projectDir, "main.go")); os.IsNotExist(err) {
		return fmt.Errorf("no main.go found")
	}
	
	return nil
}

func buildApplication(projectDir string) error {
	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(projectDir); err != nil {
		return err
	}
	
	// Run build command
	parts := strings.Fields(buildCmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if verbose {
		logStep("Running: %s", buildCmd)
	}
	
	return cmd.Run()
}

func deployToDigitalOcean(projectDir, projectName string) error {
	// Check for API key
	apiKey := os.Getenv("DIGITAL_OCEAN_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("DIGITAL_OCEAN_API_KEY environment variable is required")
	}
	
	logStep("☁️  Creating DigitalOcean droplet...")
	
	// Connect to DigitalOcean
	client := digitalocean.Connect(apiKey)
	
	// Create server configuration
	server := &digitalocean.Server{
		Name:   fmt.Sprintf("%s-prod", projectName),
		Size:   size,
		Region: region,
		Image:  "docker-20-04",
	}
	
	// Prepare deployment options
	var opts []hosting.LaunchOption
	
	// Upload application binary
	appPath := filepath.Join(projectDir, "app")
	opts = append(opts, hosting.WithFileUpload(appPath, "/usr/local/bin/app"))
	
	// Upload views directory if it exists
	viewsPath := filepath.Join(projectDir, "views")
	if _, err := os.Stat(viewsPath); err == nil {
		opts = append(opts, hosting.WithFileUpload(viewsPath, "/root/views"))
		logStep("🎨 Uploading views directory")
	}
	
	// Create systemd service and setup script
	setupScript := generateSetupScript(projectName)
	opts = append(opts, hosting.WithBinaryData("/root/setup.sh", []byte(setupScript)))
	opts = append(opts, hosting.WithSetupScript("bash", "/root/setup.sh"))
	
	// Launch server
	deployedServer, err := client.Launch(server, opts...)
	if err != nil {
		return fmt.Errorf("failed to launch server: %v", err)
	}
	
	logStep("✅ Server created successfully!")
	logStep("📍 Server ID: %s", deployedServer.GetID())
	logStep("🌍 IP Address: %s", deployedServer.GetIP())
	
	// Configure domain if provided
	if domain != "" {
		logStep("🌐 Configuring domain: %s", domain)
		if err := deployedServer.Alias("", domain); err != nil {
			logStep("⚠️  Domain configuration failed: %v", err)
		} else {
			logStep("✅ Domain configured successfully!")
		}
	}
	
	// Final instructions
	fmt.Printf("\n🎉 Deployment complete!\n\n")
	fmt.Printf("Your application is now running at:\n")
	fmt.Printf("  🔗 http://%s:5000\n", deployedServer.GetIP())
	if domain != "" {
		fmt.Printf("  🔗 https://%s\n", domain)
	}
	fmt.Printf("\n📋 Server Details:\n")
	fmt.Printf("  ID: %s\n", deployedServer.GetID())
	fmt.Printf("  IP: %s\n", deployedServer.GetIP())
	fmt.Printf("  Size: %s\n", size)
	fmt.Printf("  Region: %s\n", region)
	fmt.Printf("\n📝 To connect via SSH:\n")
	fmt.Printf("  ssh root@%s\n", deployedServer.GetIP())
	
	return nil
}

func generateSetupScript(projectName string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

echo "🚀 Setting up %s..."

# Make app executable
chmod +x /usr/local/bin/app

# Create systemd service
cat > /etc/systemd/system/%s.service << EOF
[Unit]
Description=%s Application
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root
ExecStart=/usr/local/bin/app
EnvironmentFile=-/root/.env
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
systemctl daemon-reload
systemctl enable %s
systemctl start %s

# Configure firewall
ufw allow ssh
ufw allow 5000/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo "✅ Setup complete! %s is running on port 5000"
`, projectName, projectName, projectName, projectName, projectName, projectName)
}

func logStep(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
