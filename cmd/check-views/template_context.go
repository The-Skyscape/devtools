package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// TemplateContextTracker tracks what type is passed to each template
type TemplateContextTracker struct {
	resolver      *TypeResolver
	templateMap   map[string]*ContextInfo // template name -> context type
	controllerPkg *packages.Package
}

// ContextInfo represents the context passed to a template
type ContextInfo struct {
	TemplateName string
	Type         *ResolvedType
	Source       string // "controller", "include", "inferred"
	Controller   string // Controller that renders this template
	Method       string // Method that renders this template
}

// NewTemplateContextTracker creates a new template context tracker
func NewTemplateContextTracker(resolver *TypeResolver) *TemplateContextTracker {
	return &TemplateContextTracker{
		resolver:    resolver,
		templateMap: make(map[string]*ContextInfo),
	}
}

// AnalyzeControllers analyzes all controller files to find what they pass to templates
func (tc *TemplateContextTracker) AnalyzeControllers(dir string) error {
	controllersDir := filepath.Join(dir, "controllers")

	// Load the controllers package
	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax |
			packages.NeedName |
			packages.NeedFiles,
		Dir: controllersDir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return fmt.Errorf("failed to load controllers package: %w", err)
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found in controllers directory")
	}

	tc.controllerPkg = pkgs[0]

	// Analyze each file in the package
	for _, file := range tc.controllerPkg.Syntax {
		if err := tc.analyzeFile(file); err != nil {
			if verbose {
				log.Printf("Error analyzing file: %v", err)
			}
		}
	}

	return nil
}

// analyzeFile analyzes a single controller file for Render calls
func (tc *TemplateContextTracker) analyzeFile(file *ast.File) error {
	// Find all method declarations
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			continue
		}

		// Get the receiver type (controller)
		receiverType := getReceiverType(fn.Recv)
		if !strings.HasSuffix(receiverType, "Controller") {
			continue
		}

		// Find Render calls in this method
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Look for c.Render calls
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Render" && len(call.Args) >= 3 {
					// Extract template name and data type
					if err := tc.extractRenderInfo(call, receiverType, fn.Name.Name); err != nil {
						if verbose {
							log.Printf("Failed to extract render info: %v", err)
						}
					}
				}
			}

			return true
		})
	}

	return nil
}

// extractRenderInfo extracts template name and data type from a Render call
func (tc *TemplateContextTracker) extractRenderInfo(call *ast.CallExpr, controller, method string) error {
	// Get template name from second argument (first is w, second is r, third is template)
	if len(call.Args) < 3 {
		return fmt.Errorf("Render call has too few arguments")
	}

	// Extract template name
	templateName := ""
	if lit, ok := call.Args[2].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		// Remove quotes
		templateName = strings.Trim(lit.Value, `"`)
	} else {
		return fmt.Errorf("template name is not a string literal")
	}

	// Get data type from fourth argument if present
	var dataType *ResolvedType
	if len(call.Args) >= 4 {
		dataArg := call.Args[3]

		// Try to resolve the type of the data argument
		if tc.controllerPkg != nil && tc.controllerPkg.TypesInfo != nil {
			if t := tc.controllerPkg.TypesInfo.TypeOf(dataArg); t != nil {
				dataType = tc.resolver.resolveType(t)

				// If we couldn't resolve it, try some heuristics
				if dataType == nil {
					dataType = tc.inferTypeFromExpression(dataArg)
				}
			}
		}
	}

	// Store the context info
	tc.templateMap[templateName] = &ContextInfo{
		TemplateName: templateName,
		Type:         dataType,
		Source:       "controller",
		Controller:   controller,
		Method:       method,
	}

	if verbose {
		typeName := "unknown"
		if dataType != nil {
			typeName = dataType.Name
		}
		log.Printf("Template %s rendered by %s.%s with type %s",
			templateName, controller, method, typeName)
	}

	return nil
}

// inferTypeFromExpression tries to infer the type from an expression
func (tc *TemplateContextTracker) inferTypeFromExpression(expr ast.Expr) *ResolvedType {
	switch e := expr.(type) {
	case *ast.Ident:
		// Variable name - try to guess based on naming conventions
		name := e.Name

		// Common patterns
		if strings.Contains(name, "repo") || strings.Contains(name, "Repo") {
			return tc.resolver.GetType("Repository")
		}
		if strings.Contains(name, "user") || strings.Contains(name, "User") {
			return tc.resolver.GetType("User")
		}
		if strings.Contains(name, "issue") || strings.Contains(name, "Issue") {
			return tc.resolver.GetType("Issue")
		}
		if strings.Contains(name, "workspace") || strings.Contains(name, "Workspace") {
			return tc.resolver.GetType("Workspace")
		}

	case *ast.CompositeLit:
		// Struct literal - try to get the type
		if t, ok := e.Type.(*ast.Ident); ok {
			return tc.resolver.GetType(t.Name)
		}

	case *ast.CallExpr:
		// Function call - check for common patterns
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			// Models.Get() patterns
			if sel.Sel.Name == "Get" || sel.Sel.Name == "Search" {
				if id, ok := sel.X.(*ast.Ident); ok {
					// Convert "Repositories" to "Repository"
					modelName := strings.TrimSuffix(id.Name, "ies") + "y"
					if strings.HasSuffix(id.Name, "s") && !strings.HasSuffix(id.Name, "ies") {
						modelName = strings.TrimSuffix(id.Name, "s")
					}
					return tc.resolver.GetType(modelName)
				}
			}
		}

	case *ast.UnaryExpr:
		// Pointer expressions like &Repository{}
		if e.Op == token.AND {
			return tc.inferTypeFromExpression(e.X)
		}
	}

	return nil
}

// AnalyzeTemplateIncludes analyzes template files for {{template}} directives
func (tc *TemplateContextTracker) AnalyzeTemplateIncludes(dir string) error {
	viewsDir := filepath.Join(dir, "views")

	// Parse all template files
	templates, err := ParseTemplatesWithIncludes(viewsDir)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Analyze each template's includes
	for templateName, includes := range templates {
		// Get the context for the parent template
		parentContext, exists := tc.templateMap[templateName]
		if !exists {
			// Try to infer from template name
			parentContext = tc.inferContextFromTemplateName(templateName)
		}

		// Propagate context to included templates
		if parentContext != nil && parentContext.Type != nil {
			for _, includeName := range includes {
				if _, exists := tc.templateMap[includeName]; !exists {
					tc.templateMap[includeName] = &ContextInfo{
						TemplateName: includeName,
						Type:         parentContext.Type,
						Source:       "include",
						Controller:   parentContext.Controller,
						Method:       parentContext.Method,
					}
				}
			}
		}
	}

	return nil
}

// inferContextFromTemplateName tries to guess the context type from template name
func (tc *TemplateContextTracker) inferContextFromTemplateName(templateName string) *ContextInfo {
	// Remove .html extension if present
	name := strings.TrimSuffix(templateName, ".html")

	var typeGuess *ResolvedType
	var controller string

	// Common patterns
	if strings.HasPrefix(name, "repo-") {
		typeGuess = tc.resolver.GetType("Repository")
		controller = "ReposController"
	} else if strings.HasPrefix(name, "user-") {
		typeGuess = tc.resolver.GetType("User")
		controller = "UsersController"
	} else if strings.HasPrefix(name, "issue-") {
		typeGuess = tc.resolver.GetType("Issue")
		controller = "IssuesController"
	} else if strings.HasPrefix(name, "workspace-") {
		typeGuess = tc.resolver.GetType("Workspace")
		controller = "WorkspacesController"
	} else if strings.HasPrefix(name, "admin-") {
		controller = "AdminController"
	} else if strings.HasPrefix(name, "settings-") {
		controller = "SettingsController"
	}

	if typeGuess != nil || controller != "" {
		return &ContextInfo{
			TemplateName: templateName,
			Type:         typeGuess,
			Source:       "inferred",
			Controller:   controller,
			Method:       "",
		}
	}

	return nil
}

// GetTemplateContext returns the context type for a template
func (tc *TemplateContextTracker) GetTemplateContext(templateName string) *ResolvedType {
	if ctx, exists := tc.templateMap[templateName]; exists {
		return ctx.Type
	}

	// Try to infer from template name
	if ctx := tc.inferContextFromTemplateName(templateName); ctx != nil {
		return ctx.Type
	}

	return nil
}

// GetTemplateInfo returns detailed info about a template's context
func (tc *TemplateContextTracker) GetTemplateInfo(templateName string) *ContextInfo {
	if ctx, exists := tc.templateMap[templateName]; exists {
		return ctx
	}

	// Try to infer from template name
	return tc.inferContextFromTemplateName(templateName)
}

// ParseTemplatesWithIncludes parses templates and finds {{template}} directives
func ParseTemplatesWithIncludes(dir string) (map[string][]string, error) {
	result := make(map[string][]string)

	// This would parse template files and extract {{template "name"}} directives
	// For now, returning empty map - would need template parsing logic

	return result, nil
}
