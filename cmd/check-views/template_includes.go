package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TemplateIncludeReference represents a template include statement
type TemplateIncludeReference struct {
	File         string // Template file containing the include
	Line         int    // Line number
	TemplateName string // Name used in {{template "name" .}}
	HasPath      bool   // Whether the name contains a path separator
	IsValid      bool   // Whether this is a valid include
}

// ParseTemplateIncludes finds all template include statements
func ParseTemplateIncludes(dir string) ([]TemplateIncludeReference, error) {
	var refs []TemplateIncludeReference

	// Find views directory
	viewsDir := filepath.Join(dir, "views")

	// Check if views directory exists
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		viewsDir = dir
	}

	// Walk through all HTML files
	err := filepath.WalkDir(viewsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-HTML files
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Parse the HTML file
		fileRefs, err := parseTemplateIncludesInFile(path, viewsDir)
		if err != nil {
			if verbose {
				fmt.Printf("  ⚠️  Error parsing %s: %v\n", path, err)
			}
			return nil // Continue with other files
		}

		refs = append(refs, fileRefs...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk views directory: %w", err)
	}

	return refs, nil
}

// parseTemplateIncludesInFile parses a single HTML file for template includes
func parseTemplateIncludesInFile(filePath, baseDir string) ([]TemplateIncludeReference, error) {
	var refs []TemplateIncludeReference

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get relative path for better display
	relPath, _ := filepath.Rel(baseDir, filePath)
	if relPath == "" {
		relPath = filepath.Base(filePath)
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Match {{template "name" .}} or {{template "name"}}
	// Also matches with extra whitespace
	templatePattern := regexp.MustCompile(`\{\{\s*template\s+"([^"]+)"[^}]*\}\}`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Find all template includes
		matches := templatePattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				templateName := match[1]

				// Check if it contains a path separator
				hasPath := strings.Contains(templateName, "/")

				// Determine if this is a valid include
				// "layout/start" and "layout/end" are special cases - they're valid
				// Any other path-based reference (except layout/*) is invalid
				isValid := !hasPath || strings.HasPrefix(templateName, "layout/")

				ref := TemplateIncludeReference{
					File:         relPath,
					Line:         lineNum,
					TemplateName: templateName,
					HasPath:      hasPath,
					IsValid:      isValid,
				}
				refs = append(refs, ref)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return refs, nil
}

// ValidateTemplateIncludes checks for invalid template include patterns
func ValidateTemplateIncludes(includes []TemplateIncludeReference) []ValidationError {
	var errors []ValidationError

	for _, include := range includes {
		if !include.IsValid {
			// This is a path-based template reference (not layout/*)
			// Extract the filename from the path
			parts := strings.Split(include.TemplateName, "/")
			filename := parts[len(parts)-1]

			err := ValidationError{
				File:       include.File,
				Line:       include.Line,
				Reference:  fmt.Sprintf("{{template \"%s\" .}}", include.TemplateName),
				Problem:    fmt.Sprintf("Template reference contains path '%s'", include.TemplateName),
				Suggestion: fmt.Sprintf("Use filename only: {{template \"%s\" .}}", filename),
			}
			errors = append(errors, err)
		}
	}

	return errors
}
