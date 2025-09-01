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

// ParseTemplates finds all controller references in HTML templates
func ParseTemplates(dir string) ([]TemplateReference, error) {
	var refs []TemplateReference

	// Find views directory
	viewsDir := filepath.Join(dir, "views")
	
	// Check if views directory exists
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		// Try looking for templates in the current directory
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
		fileRefs, err := parseTemplateFile(path, viewsDir)
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

// parseTemplateFile parses a single HTML template file for controller references
func parseTemplateFile(filePath, baseDir string) ([]TemplateReference, error) {
	var refs []TemplateReference

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

	// Regular expressions for different patterns
	// Match {{controller.Method}} or {{controller.Method args}}
	controllerMethodPattern := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\.\s*([A-Z][a-zA-Z0-9_]*)\s*[^}]*\}\}`)
	
	// Match {{range controller.Method}} or {{with controller.Method}}
	rangeWithPattern := regexp.MustCompile(`\{\{\s*(?:range|with)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\.\s*([A-Z][a-zA-Z0-9_]*)\s*[^}]*\}\}`)
	
	// Match {{if controller.Method}}
	ifPattern := regexp.MustCompile(`\{\{\s*if\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\.\s*([A-Z][a-zA-Z0-9_]*)\s*[^}]*\}\}`)

	// Track context for better error messages
	var currentContext []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Track {{with}} and {{range}} context
		if strings.Contains(line, "{{with") || strings.Contains(line, "{{range") {
			matches := rangeWithPattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					currentContext = append(currentContext, fmt.Sprintf("%s.%s", match[1], match[2]))
				}
			}
		}
		
		// Clear context on {{end}}
		if strings.Contains(line, "{{end") && len(currentContext) > 0 {
			currentContext = currentContext[:len(currentContext)-1]
		}

		// Find all controller.Method references
		patterns := []struct {
			re   *regexp.Regexp
			name string
		}{
			{controllerMethodPattern, "method"},
			{rangeWithPattern, "range/with"},
			{ifPattern, "if"},
		}

		for _, p := range patterns {
			matches := p.re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					controller := match[1]
					method := match[2]
					
					// Skip template helpers that aren't controllers
					if isTemplateHelper(controller) {
						continue
					}

					ref := TemplateReference{
						File:       relPath,
						Line:       lineNum,
						Controller: controller,
						Method:     method,
						Full:       fmt.Sprintf("%s.%s", controller, method),
					}
					refs = append(refs, ref)
				}
			}
		}

		// Also check for common template variables that look like controller refs
		// but might be using a different pattern
		if strings.Contains(line, "{{") && strings.Contains(line, ".") {
			// Look for patterns like {{.SomeMethod}} within a {{with controller}} block
			if len(currentContext) > 0 && regexp.MustCompile(`\{\{\s*\.\s*([A-Z][a-zA-Z0-9_]*)`).MatchString(line) {
				matches := regexp.MustCompile(`\{\{\s*\.\s*([A-Z][a-zA-Z0-9_]*)`).FindAllStringSubmatch(line, -1)
				for _, match := range matches {
					if len(match) >= 2 {
						// Use the current context controller
						contextParts := strings.Split(currentContext[len(currentContext)-1], ".")
						if len(contextParts) > 0 {
							ref := TemplateReference{
								File:       relPath,
								Line:       lineNum,
								Controller: contextParts[0],
								Method:     match[1],
								Full:       fmt.Sprintf("%s.%s", contextParts[0], match[1]),
							}
							refs = append(refs, ref)
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return refs, nil
}

// isTemplateHelper checks if a name is a built-in template helper
func isTemplateHelper(name string) bool {
	helpers := map[string]bool{
		"req":      true,
		"host":     true,
		"theme":    true,
		"path":     true,
		"path_eq":  true,
		"format":   true,
		"date":     true,
		"time":     true,
		"json":     true,
		"html":     true,
		"url":      true,
		"safe":     true,
		"slice":    true,
		"dict":     true,
		"list":     true,
		"len":      true,
		"index":    true,
		"and":      true,
		"or":       true,
		"not":      true,
		"eq":       true,
		"ne":       true,
		"lt":       true,
		"le":       true,
		"gt":       true,
		"ge":       true,
		"printf":   true,
		"print":    true,
		"println":  true,
	}
	
	return helpers[name]
}