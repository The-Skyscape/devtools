package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ConsoleReporter reports validation results to the console
type ConsoleReporter struct {
	verbose bool
	fix     bool
	quiet   bool
}

// NewReporter creates a new console reporter
func NewReporter(verbose, fix, quiet bool) Reporter {
	return &ConsoleReporter{
		verbose: verbose,
		fix:     fix,
		quiet:   quiet,
	}
}

// Report outputs the validation results
func (r *ConsoleReporter) Report(result ValidationResult) {
	if r.quiet {
		// In quiet mode, only show errors
		for _, err := range result.Errors {
			fmt.Printf("%s:%d: %s\n", err.File, err.Line, err.Problem)
			if r.fix && err.Suggestion != "" {
				fmt.Printf("  → %s\n", err.Suggestion)
			}
		}
		return
	}

	// Group errors by file for better readability
	errorsByFile := make(map[string][]ValidationError)
	for _, err := range result.Errors {
		errorsByFile[err.File] = append(errorsByFile[err.File], err)
	}

	// Sort files for consistent output
	var files []string
	for file := range errorsByFile {
		files = append(files, file)
	}
	sort.Strings(files)

	// Report errors by file
	for _, file := range files {
		errors := errorsByFile[file]
		
		// Sort errors by line number
		sort.Slice(errors, func(i, j int) bool {
			return errors[i].Line < errors[j].Line
		})

		if len(errors) == 1 {
			fmt.Printf("  ✗ %s:%d: %s not found\n", file, errors[0].Line, errors[0].Reference)
		} else {
			fmt.Printf("  ✗ %s: %d errors\n", file, len(errors))
		}

		if r.verbose || r.fix {
			for _, err := range errors {
				fmt.Printf("    Line %d: %s\n", err.Line, err.Problem)
				if err.Suggestion != "" {
					fmt.Printf("    💡 %s\n", err.Suggestion)
				}
			}
		}
	}

	// Report files that passed
	if r.verbose && len(result.Errors) == 0 {
		fmt.Printf("  ✓ All templates validated successfully\n")
	}

	// Summary
	fmt.Println()
	if result.Valid {
		fmt.Printf("✅ Success: %d references validated in %d templates\n",
			result.Summary.ValidReferences,
			result.Summary.TotalTemplates)
	} else {
		fmt.Printf("❌ Found %d errors in %d templates\n",
			result.Summary.Errors,
			len(errorsByFile))
		
		if r.fix && !r.quiet {
			fmt.Println("   Run with --fix to see suggested corrections")
		}
	}
}

// reportJSON outputs the validation result as JSON
func reportJSON(result ValidationResult) error {
	// Create a JSON-friendly structure
	output := struct {
		Valid   bool              `json:"valid"`
		Summary Statistics        `json:"summary"`
		Errors  []ValidationError `json:"errors"`
	}{
		Valid:   result.Valid,
		Summary: result.Summary,
		Errors:  result.Errors,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// Color helpers for terminal output (could be enhanced with color library)
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// colorize adds color to text if terminal supports it
func colorize(text, color string) string {
	// Check if output is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return color + text + colorReset
	}
	return text
}

// formatError formats an error message with color
func formatError(err ValidationError) string {
	location := fmt.Sprintf("%s:%d", err.File, err.Line)
	problem := err.Problem
	
	if err.Suggestion != "" {
		problem += fmt.Sprintf("\n    %s", colorize("→ "+err.Suggestion, colorCyan))
	}
	
	return fmt.Sprintf("%s: %s", 
		colorize(location, colorGray),
		colorize(problem, colorRed))
}

// DiffReporter shows a diff-like output for suggested fixes
type DiffReporter struct {
	ConsoleReporter
}

// NewDiffReporter creates a reporter that shows diff-style suggestions
func NewDiffReporter(verbose bool) Reporter {
	return &DiffReporter{
		ConsoleReporter: ConsoleReporter{
			verbose: verbose,
			fix:     true,
			quiet:   false,
		},
	}
}

// Report shows diff-style output
func (d *DiffReporter) Report(result ValidationResult) {
	if !result.Valid {
		fmt.Println("Suggested fixes:")
		fmt.Println(strings.Repeat("-", 60))
		
		for _, err := range result.Errors {
			fmt.Printf("\n%s:%d\n", err.File, err.Line)
			fmt.Printf("- %s\n", colorize(err.Reference, colorRed))
			
			if err.Suggestion != "" && strings.Contains(err.Suggestion, "Did you mean") {
				// Extract the suggested replacement
				parts := strings.Split(err.Suggestion, "'")
				if len(parts) >= 2 {
					suggested := parts[1]
					// Build the corrected reference
					refParts := strings.Split(err.Reference, ".")
					if len(refParts) == 2 {
						if strings.Contains(err.Problem, "Controller") {
							fmt.Printf("+ %s\n", colorize(suggested+"."+refParts[1], colorGreen))
						} else {
							fmt.Printf("+ %s\n", colorize(refParts[0]+"."+suggested, colorGreen))
						}
					}
				}
			}
		}
		
		fmt.Println(strings.Repeat("-", 60))
	}
	
	// Call parent reporter for summary
	d.ConsoleReporter.Report(result)
}