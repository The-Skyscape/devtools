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
		fmt.Printf("  ✓ Found %d controllers with %d methods\n", len(controllers), totalMethods)
	}

	if !quiet && !jsonOut {
		fmt.Println("\n🔍 Discovering types (models, internal, templates)...")
	}

	// Discover all types (models, internal structs, controller returns)
	allTypes, err := DiscoverAllTypes(dir)
	if err != nil {
		return fmt.Errorf("failed to discover types: %w", err)
	}

	// For backward compatibility, extract just models
	models := make(map[string]*ModelInfo)
	for name, typeInfo := range allTypes {
		if typeInfo.Source == "model" {
			models[name] = &ModelInfo{
				Name:     typeInfo.Name,
				Package:  typeInfo.Package,
				FilePath: typeInfo.FilePath,
				Fields:   typeInfo.Fields,
				Methods:  typeInfo.Methods,
			}
		}
	}

	if !quiet && !jsonOut {
		totalFields := 0
		modelCount := 0
		internalCount := 0
		for _, t := range allTypes {
			totalFields += len(t.Fields)
			switch t.Source {
			case "model":
				modelCount++
			case "internal":
				internalCount++
			}
		}
		fmt.Printf("  ✓ Found %d types with %d fields\n", len(allTypes), totalFields)
		fmt.Printf("    - %d models\n", modelCount)
		fmt.Printf("    - %d internal types\n", internalCount)
	}

	if verbose && !jsonOut {
		fmt.Println("\nControllers:")
		for _, c := range controllers {
			fmt.Printf("  %s (%s)\n", c.Prefix, c.Type)
			fmt.Printf("    File: %s\n", c.FilePath)
			fmt.Printf("    Methods: %d\n", len(c.Methods))
			if verbose {
				for _, m := range c.Methods {
					fmt.Printf("      - %s\n", m)
				}
			}
		}
		
		fmt.Println("\nModels:")
		for name, m := range models {
			fmt.Printf("  %s\n", name)
			fmt.Printf("    File: %s\n", m.FilePath)
			fmt.Printf("    Fields: %d\n", len(m.Fields))
			fmt.Printf("    Methods: %d\n", len(m.Methods))
		}
		fmt.Println()
	}

	if !quiet && !jsonOut {
		fmt.Println("\n🛣️  Discovering routes...")
	}

	// Discover routes
	routes, err := DiscoverRoutes(dir)
	if err != nil {
		return fmt.Errorf("failed to discover routes: %w", err)
	}

	if !quiet && !jsonOut {
		fmt.Printf("  ✓ Found %d routes\n", len(routes))
	}

	if !quiet && !jsonOut {
		fmt.Println("\n📋 Parsing templates...")
	}

	// Parse templates with both methods
	// First use regex-based parser for controller references
	regexRefs, err := ParseTemplates(dir)
	if err != nil {
		return fmt.Errorf("failed to parse templates with regex: %w", err)
	}
	
	// Then use AST parser for enhanced field validation
	templateRefs, fieldRefs, err := ParseTemplatesWithAST(dir, controllers, allTypes)
	if err != nil {
		return fmt.Errorf("failed to parse templates with AST: %w", err)
	}
	
	// Combine regex refs with AST refs (regex is more reliable for controller refs)
	if len(regexRefs) > 0 {
		templateRefs = regexRefs
	}

	// Parse URL references
	urlRefs, err := ParseHostReferences(dir)
	if err != nil {
		return fmt.Errorf("failed to parse URL references: %w", err)
	}

	if !quiet && !jsonOut {
		fmt.Printf("  ✓ Found %d controller references\n", len(templateRefs))
		fmt.Printf("  ✓ Found %d field references\n", len(fieldRefs))
		fmt.Printf("  ✓ Found %d URL references\n", len(urlRefs))
	}

	// Validate URL references
	urlErrors := ValidateURLReferences(urlRefs, routes)

	// Validate all references using enhanced validation with all types
	result := ValidateWithTypes(controllers, allTypes, templateRefs, fieldRefs)
	
	// Add URL errors to result
	result.URLErrors = urlErrors
	result.Summary.TotalURLRefs = len(urlRefs)
	result.Summary.ValidURLRefs = len(urlRefs) - len(urlErrors)
	result.Summary.URLErrors = len(urlErrors)
	
	// Update valid flag if there are URL errors
	if len(urlErrors) > 0 {
		result.Valid = false
	}

	// Report results
	if jsonOut {
		return reportEnhancedJSON(*result)
	}

	reporter := NewEnhancedReporter(verbose, fix, quiet)
	reporter.Report(*result)

	totalErrors := len(result.ControllerErrors) + len(result.FieldErrors) + len(result.URLErrors)
	if !result.Valid {
		return fmt.Errorf("validation failed with %d errors", totalErrors)
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

// Enhanced types for better validation
type EnhancedValidationResult struct {
	Valid            bool
	ControllerErrors []ValidationError
	FieldErrors      []FieldValidationError
	URLErrors        []URLValidationError
	Summary          EnhancedStatistics
}

type FieldValidationError struct {
	File       string
	Line       int
	Expression string
	Fields     []string
	Context    string
	Problem    string
	Suggestion string
}

type EnhancedStatistics struct {
	TotalTemplates      int
	TotalControllerRefs int
	ValidControllerRefs int
	TotalFieldRefs      int
	ValidFieldRefs      int
	TotalURLRefs        int
	ValidURLRefs        int
	ControllerErrors    int
	FieldErrors         int
	URLErrors           int
}

type Reporter interface {
	Report(ValidationResult)
}