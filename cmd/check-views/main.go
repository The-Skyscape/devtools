package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	fix     bool
	config  string
	jsonOut bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "check-views [directory]",
	Short: "Validate template references to controller methods",
	Long: `Automatically discovers controllers and validates all template references.

This tool helps catch template errors after refactoring by:
  - Auto-discovering controller mappings from factory functions
  - Parsing all HTML templates for controller method references
  - Validating that referenced methods exist
  - Providing smart suggestions for typos

Examples:
  check-views                    # Check current directory
  check-views ../website         # Check specific directory
  check-views --verbose          # Show detailed output
  check-views --json            # Output JSON for CI/CD`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidation,
}

func init() {
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	rootCmd.Flags().BoolVarP(&fix, "fix", "f", false, "Suggest fixes for common issues")
	rootCmd.Flags().StringVarP(&config, "config", "c", "", "Path to config file")
	rootCmd.Flags().BoolVarP(&jsonOut, "json", "j", false, "Output as JSON for CI/CD")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only show errors")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runValidation(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	if !quiet && !jsonOut {
		fmt.Println("✨ Discovering controllers...")
	}

	// Discover controllers
	controllers, err := DiscoverControllers(dir)
	if err != nil {
		return fmt.Errorf("failed to discover controllers: %w", err)
	}

	if !quiet && !jsonOut {
		totalMethods := 0
		for _, c := range controllers {
			totalMethods += len(c.Methods)
		}
		fmt.Printf("  ✓ Found %d controllers with %d methods\n\n", len(controllers), totalMethods)
	}

	if verbose && !jsonOut {
		for _, c := range controllers {
			fmt.Printf("  Controller: %s (%s)\n", c.Prefix, c.Type)
			fmt.Printf("    File: %s\n", c.FilePath)
			fmt.Printf("    Methods: %d\n", len(c.Methods))
			if verbose {
				for _, m := range c.Methods {
					fmt.Printf("      - %s\n", m)
				}
			}
		}
		fmt.Println()
	}

	if !quiet && !jsonOut {
		fmt.Println("📋 Checking templates...")
	}

	// Parse templates
	refs, err := ParseTemplates(dir)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Validate references
	result := Validate(controllers, refs)

	// Report results
	if jsonOut {
		return reportJSON(result)
	}

	reporter := NewReporter(verbose, fix, quiet)
	reporter.Report(result)

	if !result.Valid {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}

// Functions are implemented in their respective files:
// - DiscoverControllers in discovery.go
// - ParseTemplates in parser.go
// - Validate in validator.go
// - reportJSON in reporter.go

// Type definitions that will be moved to their respective files
type ControllerInfo struct {
	Prefix   string   // e.g., "admin"
	Type     string   // e.g., "AdminController"
	FilePath string   // e.g., "controllers/admin.go"
	Methods  []string // Public methods found
}

type TemplateReference struct {
	File       string // e.g., "admin-dashboard.html"
	Line       int    // Line number
	Controller string // e.g., "admin"
	Method     string // e.g., "GetTotalUsers"
	Full       string // e.g., "admin.GetTotalUsers"
}

type ValidationResult struct {
	Valid   bool
	Errors  []ValidationError
	Summary Statistics
}

type ValidationError struct {
	File       string
	Line       int
	Reference  string
	Problem    string
	Suggestion string // e.g., "Did you mean GetUsers?"
}

type Statistics struct {
	TotalTemplates  int
	TotalReferences int
	ValidReferences int
	Errors          int
}

type Reporter interface {
	Report(ValidationResult)
}