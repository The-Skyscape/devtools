package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// EnhancedReporter handles reporting for both controller and field validation
type EnhancedReporter struct {
	verbose bool
	fix     bool
	quiet   bool
}

// NewEnhancedReporter creates a new enhanced reporter
func NewEnhancedReporter(verbose, fix, quiet bool) *EnhancedReporter {
	return &EnhancedReporter{
		verbose: verbose,
		fix:     fix,
		quiet:   quiet,
	}
}

// Report outputs the enhanced validation results
func (r *EnhancedReporter) Report(result EnhancedValidationResult) {
	if r.quiet {
		// In quiet mode, only show errors
		for _, err := range result.ControllerErrors {
			fmt.Printf("%s:%d: %s\n", err.File, err.Line, err.Problem)
			if r.fix && err.Suggestion != "" {
				fmt.Printf("  → %s\n", err.Suggestion)
			}
		}
		for _, err := range result.FieldErrors {
			fmt.Printf("%s:%d: %s\n", err.File, err.Line, err.Problem)
			if r.fix && err.Suggestion != "" {
				fmt.Printf("  → %s\n", err.Suggestion)
			}
		}
		for _, err := range result.URLErrors {
			fmt.Printf("%s:%d: %s\n", err.File, err.Line, err.Problem)
			if r.fix && err.Suggestion != "" {
				fmt.Printf("  → %s\n", err.Suggestion)
			}
		}
		return
	}

	// Report controller errors
	if len(result.ControllerErrors) > 0 {
		r.reportControllerErrors(result.ControllerErrors)
	}

	// Report field errors
	if len(result.FieldErrors) > 0 {
		r.reportFieldErrors(result.FieldErrors)
	}

	// Report URL errors
	if len(result.URLErrors) > 0 {
		r.reportURLErrors(result.URLErrors)
	}

	// Report summary
	r.reportSummary(result)
}

// reportControllerErrors reports controller validation errors
func (r *EnhancedReporter) reportControllerErrors(errors []ValidationError) {
	if !r.quiet {
		fmt.Println("\n🎮 Controller Reference Errors:")
	}

	// Group errors by file
	errorsByFile := make(map[string][]ValidationError)
	for _, err := range errors {
		errorsByFile[err.File] = append(errorsByFile[err.File], err)
	}

	// Sort files
	var files []string
	for file := range errorsByFile {
		files = append(files, file)
	}
	sort.Strings(files)

	// Report errors by file
	for _, file := range files {
		fileErrors := errorsByFile[file]
		
		// Sort errors by line number
		sort.Slice(fileErrors, func(i, j int) bool {
			return fileErrors[i].Line < fileErrors[j].Line
		})

		fmt.Printf("  ✗ %s: %d controller errors\n", file, len(fileErrors))

		if r.verbose || r.fix {
			for _, err := range fileErrors {
				fmt.Printf("    Line %d: %s\n", err.Line, err.Problem)
				if err.Suggestion != "" {
					fmt.Printf("    💡 %s\n", err.Suggestion)
				}
			}
		}
	}
}

// reportFieldErrors reports field validation errors
func (r *EnhancedReporter) reportFieldErrors(errors []FieldValidationError) {
	if !r.quiet {
		fmt.Println("\n📦 Model Field Errors:")
	}

	// Group errors by file
	errorsByFile := make(map[string][]FieldValidationError)
	for _, err := range errors {
		errorsByFile[err.File] = append(errorsByFile[err.File], err)
	}

	// Sort files
	var files []string
	for file := range errorsByFile {
		files = append(files, file)
	}
	sort.Strings(files)

	// Report errors by file
	for _, file := range files {
		fileErrors := errorsByFile[file]
		
		// Sort errors by line number
		sort.Slice(fileErrors, func(i, j int) bool {
			return fileErrors[i].Line < fileErrors[j].Line
		})

		fmt.Printf("  ✗ %s: %d field errors\n", file, len(fileErrors))

		if r.verbose || r.fix {
			for _, err := range fileErrors {
				fmt.Printf("    Line %d: Expression '%s'\n", err.Line, err.Expression)
				fmt.Printf("      Context: %s\n", err.Context)
				fmt.Printf("      Problem: %s\n", err.Problem)
				if err.Suggestion != "" {
					fmt.Printf("      💡 %s\n", err.Suggestion)
				}
			}
		}
	}
}

// reportURLErrors reports URL validation errors
func (r *EnhancedReporter) reportURLErrors(errors []URLValidationError) {
	if !r.quiet {
		fmt.Println("\n🔗 URL Reference Errors:")
	}

	// Group errors by file
	errorsByFile := make(map[string][]URLValidationError)
	for _, err := range errors {
		errorsByFile[err.File] = append(errorsByFile[err.File], err)
	}

	// Sort files
	var files []string
	for file := range errorsByFile {
		files = append(files, file)
	}
	sort.Strings(files)

	// Report errors by file
	for _, file := range files {
		fileErrors := errorsByFile[file]
		
		// Sort errors by line number
		sort.Slice(fileErrors, func(i, j int) bool {
			return fileErrors[i].Line < fileErrors[j].Line
		})

		fmt.Printf("  ✗ %s: %d URL errors\n", file, len(fileErrors))

		if r.verbose || r.fix {
			for _, err := range fileErrors {
				fmt.Printf("    Line %d: URL '%s'\n", err.Line, err.URL)
				fmt.Printf("      Problem: %s\n", err.Problem)
				if err.Suggestion != "" {
					fmt.Printf("      💡 %s\n", err.Suggestion)
				}
			}
		}
	}
}

// reportSummary reports the validation summary
func (r *EnhancedReporter) reportSummary(result EnhancedValidationResult) {
	fmt.Println()
	
	if result.Valid {
		fmt.Printf("✅ Success: All references validated successfully\n")
		fmt.Printf("   • %d controller references validated\n", result.Summary.ValidControllerRefs)
		fmt.Printf("   • %d field references validated\n", result.Summary.ValidFieldRefs)
		fmt.Printf("   • %d URL references validated\n", result.Summary.ValidURLRefs)
		fmt.Printf("   • %d templates checked\n", result.Summary.TotalTemplates)
	} else {
		totalErrors := result.Summary.ControllerErrors + result.Summary.FieldErrors + result.Summary.URLErrors
		fmt.Printf("❌ Found %d errors\n", totalErrors)
		
		if result.Summary.ControllerErrors > 0 {
			fmt.Printf("   • %d controller reference errors\n", result.Summary.ControllerErrors)
		}
		if result.Summary.FieldErrors > 0 {
			fmt.Printf("   • %d field reference errors\n", result.Summary.FieldErrors)
		}
		if result.Summary.URLErrors > 0 {
			fmt.Printf("   • %d URL reference errors\n", result.Summary.URLErrors)
		}
		
		fmt.Printf("   • %d/%d controller references valid\n", 
			result.Summary.ValidControllerRefs, result.Summary.TotalControllerRefs)
		fmt.Printf("   • %d/%d field references valid\n", 
			result.Summary.ValidFieldRefs, result.Summary.TotalFieldRefs)
		fmt.Printf("   • %d/%d URL references valid\n",
			result.Summary.ValidURLRefs, result.Summary.TotalURLRefs)
		
		if !r.quiet && !r.fix {
			fmt.Println("\n   💡 Run with --fix to see suggested corrections")
		}
	}
}

// reportEnhancedJSON outputs the enhanced validation result as JSON
func reportEnhancedJSON(result EnhancedValidationResult) error {
	// Create a JSON-friendly structure
	output := struct {
		Valid            bool                     `json:"valid"`
		Summary          EnhancedStatistics       `json:"summary"`
		ControllerErrors []ValidationError        `json:"controller_errors"`
		FieldErrors      []FieldValidationError   `json:"field_errors"`
		URLErrors        []URLValidationError     `json:"url_errors"`
	}{
		Valid:            result.Valid,
		Summary:          result.Summary,
		ControllerErrors: result.ControllerErrors,
		FieldErrors:      result.FieldErrors,
		URLErrors:        result.URLErrors,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// EnhancedDiffReporter shows diff-style output for both types of errors
type EnhancedDiffReporter struct {
	EnhancedReporter
}

// NewEnhancedDiffReporter creates a diff-style reporter
func NewEnhancedDiffReporter(verbose bool) *EnhancedDiffReporter {
	return &EnhancedDiffReporter{
		EnhancedReporter: EnhancedReporter{
			verbose: verbose,
			fix:     true,
			quiet:   false,
		},
	}
}

// Report shows diff-style output for enhanced results
func (d *EnhancedDiffReporter) Report(result EnhancedValidationResult) {
	if !result.Valid {
		fmt.Println("Suggested fixes:")
		fmt.Println(strings.Repeat("-", 60))
		
		// Show controller fixes
		if len(result.ControllerErrors) > 0 {
			fmt.Println("\nController Reference Fixes:")
			for _, err := range result.ControllerErrors {
				d.showControllerFix(err)
			}
		}
		
		// Show field fixes
		if len(result.FieldErrors) > 0 {
			fmt.Println("\nModel Field Fixes:")
			for _, err := range result.FieldErrors {
				d.showFieldFix(err)
			}
		}
		
		fmt.Println(strings.Repeat("-", 60))
	}
	
	// Call parent reporter for summary
	d.EnhancedReporter.Report(result)
}

// showControllerFix shows a fix for a controller error
func (d *EnhancedDiffReporter) showControllerFix(err ValidationError) {
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

// showFieldFix shows a fix for a field error
func (d *EnhancedDiffReporter) showFieldFix(err FieldValidationError) {
	fmt.Printf("\n%s:%d\n", err.File, err.Line)
	fmt.Printf("- %s\n", colorize(err.Expression, colorRed))
	
	if err.Suggestion != "" && strings.Contains(err.Suggestion, "Did you mean") {
		// Extract the suggested replacement
		parts := strings.Split(err.Suggestion, "'")
		if len(parts) >= 2 {
			suggested := parts[1]
			// Replace the last field in the expression with the suggestion
			exprParts := strings.Split(err.Expression, ".")
			if len(exprParts) > 1 {
				exprParts[len(exprParts)-1] = suggested
				fmt.Printf("+ %s\n", colorize(strings.Join(exprParts, "."), colorGreen))
			}
		}
	}
}