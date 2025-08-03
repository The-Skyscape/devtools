package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates
var templates embed.FS

var (
	force bool
)

type ProjectData struct {
	Name       string
	ModuleName string
}

var rootCmd = &cobra.Command{
	Use:   "create-app [project-name]",
	Short: "Create a new TheSkyscape DevTools todo application",
	Long: `Create a new TheSkyscape DevTools todo application with authentication, HTMX, and modern UI.

The generated application includes:
  - User authentication with signup/signin
  - Todo CRUD operations with priorities and due dates
  - Real-time updates with HTMX
  - Modern UI with DaisyUI components
  - SQLite database with dynamic ORM

Examples:
  create-app my-todo-app
  create-app my-project --force`,
	Args: cobra.ExactArgs(1),
	Run:  createApp,
}

func init() {
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation even if directory exists")
}

func createApp(cmd *cobra.Command, args []string) {
	projectName := args[0]
	
	// Check if directory exists
	if _, err := os.Stat(projectName); err == nil && !force {
		fmt.Fprintf(os.Stderr, "Error: Directory '%s' already exists. Use --force to overwrite.\n", projectName)
		os.Exit(1)
	}
	
	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}
	
	// Prepare template data
	data := ProjectData{
		Name:       projectName,
		ModuleName: strings.ReplaceAll(projectName, "-", "_"),
	}
	
	// Generate project files
	if err := generateProject(projectName, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating project: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Successfully created '%s' todo application\n\n", projectName)
	fmt.Printf("Get started:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  go run .\n\n")
	fmt.Printf("Visit http://localhost:8080 to see your todo application!\n")
}

func generateProject(projectPath string, data ProjectData) error {
	templatePath := "templates"
	
	return fs.WalkDir(templates, templatePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the template root directory
		if path == templatePath {
			return nil
		}
		
		// Calculate relative path from template root
		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return err
		}
		
		// Target path in project
		targetPath := filepath.Join(projectPath, relPath)
		
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}
		
		// Read template file
		content, err := templates.ReadFile(path)
		if err != nil {
			return err
		}
		
		// Process template if it's a .tmpl file
		if strings.HasSuffix(path, ".tmpl") {
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")
			
			tmpl, err := template.New("file").Parse(string(content))
			if err != nil {
				return err
			}
			
			file, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer file.Close()
			
			return tmpl.Execute(file, data)
		} else {
			// Copy file as-is
			return os.WriteFile(targetPath, content, 0644)
		}
	})
}


func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
