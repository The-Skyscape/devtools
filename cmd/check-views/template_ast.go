package main

import (
	"os"
	"path/filepath"
	"strings"
	"text/template/parse"
)

// TemplateContext represents the current context in a template
type TemplateContext struct {
	Type       string   // Current type in context (e.g., "Workspace", "User")
	Variable   string   // Variable name if in a range/with
	Parent     *TemplateContext // Parent context for nested with/range
}

// FieldReference represents a field access in a template
type FieldReference struct {
	File       string   // Template file
	Line       int      // Line number (approximate)
	Expression string   // Full expression (e.g., ".User.Name")
	Fields     []string // Field chain (e.g., ["User", "Name"])
	Context    string   // Context type where this appears
}

// ParseTemplatesWithAST parses templates using Go's template parser
func ParseTemplatesWithAST(dir string, controllers []ControllerInfo, models map[string]*ModelInfo) ([]TemplateReference, []FieldReference, error) {
	var templateRefs []TemplateReference
	var fieldRefs []FieldReference
	
	// Find views directory
	viewsDir := filepath.Join(dir, "views")
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		viewsDir = dir
	}
	
	// Walk through all HTML files
	err := filepath.WalkDir(viewsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-HTML files
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		
		// Parse the template file
		refs, fields, err := parseTemplateAST(path, viewsDir, controllers, models)
		if err != nil {
			// Continue with other files even if one fails
			return nil
		}
		
		templateRefs = append(templateRefs, refs...)
		fieldRefs = append(fieldRefs, fields...)
		
		return nil
	})
	
	return templateRefs, fieldRefs, err
}

// parseTemplateAST parses a single template file using the AST
func parseTemplateAST(filePath, baseDir string, controllers []ControllerInfo, models map[string]*ModelInfo) ([]TemplateReference, []FieldReference, error) {
	var templateRefs []TemplateReference
	var fieldRefs []FieldReference
	
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	
	// Get relative path for better display
	relPath, _ := filepath.Rel(baseDir, filePath)
	if relPath == "" {
		relPath = filepath.Base(filePath)
	}
	
	// Parse the template
	tree := parse.New(relPath)
	_, err = tree.Parse(string(content), "{{", "}}", make(map[string]*parse.Tree))
	if err != nil {
		// Template parse error - might be malformed
		return templateRefs, fieldRefs, nil
	}
	
	// Process the tree
	refs, fields := parseTemplateTree(tree, relPath, controllers, models)
	templateRefs = append(templateRefs, refs...)
	fieldRefs = append(fieldRefs, fields...)
	
	return templateRefs, fieldRefs, nil
}

// isControllerName checks if a name matches a known controller prefix
func isControllerName(name string, controllers []ControllerInfo) bool {
	for _, ctrl := range controllers {
		if ctrl.Prefix == name {
			return true
		}
	}
	return false
}