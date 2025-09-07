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
	// Only match controller.Method patterns where controller starts with lowercase
	// and Method starts with uppercase (Go convention)
	// This excludes .Property patterns which are object properties

	// Match {{controller.Method}} or {{controller.Method args}} but NOT {{.Property}}
	controllerMethodPattern := regexp.MustCompile(`\{\{\s*([a-z][a-zA-Z0-9_]*)\s*\.\s*([A-Z][a-zA-Z0-9_]*)\s*[^}]*\}\}`)

	// Match {{range controller.Method}} or {{with controller.Method}}
	rangeWithPattern := regexp.MustCompile(`\{\{\s*(?:range|with)\s+([a-z][a-zA-Z0-9_]*)\s*\.\s*([A-Z][a-zA-Z0-9_]*)\s*[^}]*\}\}`)

	// Match {{if controller.Method}} - but only if it's actually a controller reference
	ifPattern := regexp.MustCompile(`\{\{\s*if\s+([a-z][a-zA-Z0-9_]*)\s*\.\s*([A-Z][a-zA-Z0-9_]*)\s*[^}]*\}\}`)

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

		// Don't try to parse .Property patterns - those are object properties, not controller methods
		// Controller references are always fully qualified as controller.Method
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return refs, nil
}

// isTemplateHelper checks if a name is a built-in template helper or keyword
func isTemplateHelper(name string) bool {
	helpers := map[string]bool{
		// Template keywords
		"if":       true,
		"else":     true,
		"end":      true,
		"range":    true,
		"with":     true,
		"template": true,
		"define":   true,
		"block":    true,

		// Core template functions
		"req":     true,
		"host":    true,
		"theme":   true,
		"path":    true,
		"path_eq": true,
		"title":   true,
		"prefix":  true,

		// Math functions
		"add":   true,
		"sub":   true,
		"mul":   true,
		"div":   true,
		"float": true,

		// String functions
		"toString":  true,
		"slice":     true,
		"head":      true,
		"default":   true,
		"set":       true,
		"dict":      true,
		"hasPrefix": true,
		"hasSuffix": true,
		"contains":  true,
		"replace":   true,
		"lower":     true,
		"upper":     true,
		"trim":      true,
		"split":     true,
		"join":      true,
		"truncate":  true,

		// Formatting functions (from FormatHelpers)
		"formatBytes":    true,
		"formatPrice":    true,
		"formatPercent":  true,
		"formatDuration": true,
		"formatDate":     true,
		"formatDateTime": true,
		"formatNumber":   true,
		"pluralize":      true,

		// JSON functions
		"jsonify": true,

		// Chart functions
		"renderChart":      true,
		"renderSparkline":  true,
		"placeholderChart": true,
		"chartLoader":      true,

		// Standard template functions
		"len":     true,
		"index":   true,
		"and":     true,
		"or":      true,
		"not":     true,
		"eq":      true,
		"ne":      true,
		"lt":      true,
		"le":      true,
		"gt":      true,
		"ge":      true,
		"printf":  true,
		"print":   true,
		"println": true,
		"html":    true,
		"js":      true,
		"url":     true,
		"call":    true,
	}

	return helpers[name]
}
