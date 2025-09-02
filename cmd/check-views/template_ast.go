package main

import (
	"fmt"
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
			if verbose {
				fmt.Printf("  ⚠️  Error parsing template %s: %v\n", path, err)
			}
			return nil // Continue with other files
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
	if tree != nil && tree.Root != nil {
		// Walk the AST
		context := &TemplateContext{Type: "root"}
		walkNode(tree.Root, relPath, context, controllers, models, &templateRefs, &fieldRefs)
	}
	
	return templateRefs, fieldRefs, nil
}

// walkNode recursively walks the template AST
func walkNode(node parse.Node, file string, context *TemplateContext, controllers []ControllerInfo, models map[string]*ModelInfo, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if node == nil {
		return
	}
	
	switch n := node.(type) {
	case *parse.ActionNode:
		if n.Pipe != nil {
			processPipe(n.Pipe, file, n.Line, context, controllers, models, templateRefs, fieldRefs)
		}
		
	case *parse.IfNode:
		if n.Pipe != nil {
			processPipe(n.Pipe, file, n.Line, context, controllers, models, templateRefs, fieldRefs)
		}
		if n.List != nil {
			walkNode(n.List, file, context, controllers, models, templateRefs, fieldRefs)
		}
		if n.ElseList != nil {
			walkNode(n.ElseList, file, context, controllers, models, templateRefs, fieldRefs)
		}
		
	case *parse.RangeNode:
		// Range changes context
		newContext := context
		if n.Pipe != nil {
			// Try to determine the type being ranged over
			if contextType := extractRangeType(n.Pipe, controllers, models); contextType != "" {
				newContext = &TemplateContext{
					Type:   contextType,
					Parent: context,
				}
			}
			processPipe(n.Pipe, file, n.Line, context, controllers, models, templateRefs, fieldRefs)
		}
		if n.List != nil {
			walkNode(n.List, file, newContext, controllers, models, templateRefs, fieldRefs)
		}
		if n.ElseList != nil {
			walkNode(n.ElseList, file, context, controllers, models, templateRefs, fieldRefs)
		}
		
	case *parse.WithNode:
		// With changes context
		newContext := context
		if n.Pipe != nil {
			// Try to determine the type in the with
			if contextType := extractWithType(n.Pipe, controllers, models); contextType != "" {
				newContext = &TemplateContext{
					Type:   contextType,
					Parent: context,
				}
			}
			processPipe(n.Pipe, file, n.Line, context, controllers, models, templateRefs, fieldRefs)
		}
		if n.List != nil {
			walkNode(n.List, file, newContext, controllers, models, templateRefs, fieldRefs)
		}
		if n.ElseList != nil {
			walkNode(n.ElseList, file, context, controllers, models, templateRefs, fieldRefs)
		}
		
	case *parse.ListNode:
		if n != nil && n.Nodes != nil {
			for _, child := range n.Nodes {
				walkNode(child, file, context, controllers, models, templateRefs, fieldRefs)
			}
		}
		
	case *parse.TextNode:
		// Text nodes don't contain template expressions
		
	default:
		// Handle other node types if needed
	}
}

// processPipe processes a pipe command for references
func processPipe(pipe *parse.PipeNode, file string, line int, context *TemplateContext, controllers []ControllerInfo, models map[string]*ModelInfo, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if pipe == nil {
		return
	}
	
	for _, cmd := range pipe.Cmds {
		processCommand(cmd, file, line, context, controllers, models, templateRefs, fieldRefs)
	}
}

// processCommand processes a single command for references
func processCommand(cmd *parse.CommandNode, file string, line int, context *TemplateContext, controllers []ControllerInfo, models map[string]*ModelInfo, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if cmd == nil || len(cmd.Args) == 0 {
		return
	}
	
	// Process each argument
	for _, arg := range cmd.Args {
		switch a := arg.(type) {
		case *parse.FieldNode:
			// Field access like .Name or .User.Email
			if len(a.Ident) > 0 {
				// Check if it's a controller reference (first ident is lowercase controller name)
				if len(a.Ident) >= 2 && isControllerName(a.Ident[0], controllers) {
					// Controller reference
					ref := TemplateReference{
						File:       file,
						Line:       line,
						Controller: a.Ident[0],
						Method:     a.Ident[1],
						Full:       fmt.Sprintf("%s.%s", a.Ident[0], a.Ident[1]),
					}
					*templateRefs = append(*templateRefs, ref)
					
					// If there are more fields, it's accessing a field on the result
					if len(a.Ident) > 2 {
						remainingFields := a.Ident[2:]
						fieldRef := FieldReference{
							File:       file,
							Line:       line,
							Expression: "." + strings.Join(remainingFields, "."),
							Fields:     remainingFields,
							Context:    "controller_result",
						}
						*fieldRefs = append(*fieldRefs, fieldRef)
					}
				} else {
					// Model field access
					fieldRef := FieldReference{
						File:       file,
						Line:       line,
						Expression: "." + strings.Join(a.Ident, "."),
						Fields:     a.Ident,
						Context:    context.Type,
					}
					*fieldRefs = append(*fieldRefs, fieldRef)
				}
			}
			
		case *parse.VariableNode:
			// Variable like $var - usually from range
			// Could track these for more advanced validation
			
		case *parse.IdentifierNode:
			// Function call or keyword
			// Skip template functions
			
		default:
			// Recursively process nested nodes if applicable
		}
	}
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

// extractRangeType tries to determine the type being ranged over
func extractRangeType(pipe *parse.PipeNode, controllers []ControllerInfo, models map[string]*ModelInfo) string {
	if pipe == nil || len(pipe.Cmds) == 0 {
		return ""
	}
	
	// Look at the first command
	cmd := pipe.Cmds[0]
	if len(cmd.Args) == 0 {
		return ""
	}
	
	// Check if it's a field access
	if field, ok := cmd.Args[0].(*parse.FieldNode); ok && len(field.Ident) >= 2 {
		// Check if it's a controller method
		for _, ctrl := range controllers {
			if ctrl.Prefix == field.Ident[0] {
				// It's a controller method - try to infer return type
				methodName := field.Ident[1]
				return inferMethodReturnType(methodName)
			}
		}
	}
	
	return ""
}

// extractWithType tries to determine the type in a with statement
func extractWithType(pipe *parse.PipeNode, controllers []ControllerInfo, models map[string]*ModelInfo) string {
	if pipe == nil || len(pipe.Cmds) == 0 {
		return ""
	}
	
	// Look at the first command
	cmd := pipe.Cmds[0]
	if len(cmd.Args) == 0 {
		return ""
	}
	
	// Check if it's a field access
	if field, ok := cmd.Args[0].(*parse.FieldNode); ok && len(field.Ident) >= 2 {
		// Check if it's a controller method
		for _, ctrl := range controllers {
			if ctrl.Prefix == field.Ident[0] {
				// It's a controller method - try to infer return type
				methodName := field.Ident[1]
				return inferMethodReturnType(methodName)
			}
		}
		
		// Check if it's a model field access
		if len(field.Ident) == 1 {
			fieldName := field.Ident[0]
			// Try to find the field type in known models
			for _, model := range models {
				if fieldInfo, exists := model.Fields[fieldName]; exists {
					return fieldInfo.Type
				}
			}
		}
	}
	
	return ""
}

// inferMethodReturnType tries to guess the return type based on method name
func inferMethodReturnType(methodName string) string {
	// Common patterns
	switch {
	case strings.HasPrefix(methodName, "Get") && strings.Contains(methodName, "User"):
		return "User"
	case strings.HasPrefix(methodName, "Get") && strings.Contains(methodName, "Workspace"):
		return "Workspace"
	case strings.HasPrefix(methodName, "Get") && strings.Contains(methodName, "Org"):
		return "Organization"
	case strings.HasPrefix(methodName, "Current") && strings.Contains(methodName, "User"):
		return "User"
	case strings.HasPrefix(methodName, "Current") && strings.Contains(methodName, "Workspace"):
		return "Workspace"
	case strings.HasPrefix(methodName, "All") || strings.HasPrefix(methodName, "List"):
		// Returns a slice - try to infer element type
		if strings.Contains(methodName, "User") {
			return "User"
		}
		if strings.Contains(methodName, "Workspace") {
			return "Workspace"
		}
		if strings.Contains(methodName, "Org") {
			return "Organization"
		}
	}
	
	return ""
}