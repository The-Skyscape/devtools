package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

// DiscoverControllers automatically finds all controllers in the given directory
func DiscoverControllers(dir string) ([]ControllerInfo, error) {
	var controllers []ControllerInfo

	// Find controllers directory
	controllersDir := filepath.Join(dir, "controllers")
	
	// Walk through all Go files in controllers directory
	err := filepath.WalkDir(controllersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse the Go file
		fileControllers, err := parseControllerFile(path)
		if err != nil {
			if verbose {
				fmt.Printf("  ⚠️  Skipping %s: %v\n", path, err)
			}
			return nil // Continue with other files
		}

		controllers = append(controllers, fileControllers...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk controllers directory: %w", err)
	}

	// Add special controllers from devtools
	controllers = append(controllers, getSpecialControllers()...)

	return controllers, nil
}

// parseControllerFile parses a single Go file for controller definitions
func parseControllerFile(filePath string) ([]ControllerInfo, error) {
	var controllers []ControllerInfo

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Map to store controller types and their methods
	controllerTypes := make(map[string]*ControllerInfo)

	// First pass: Find factory functions
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil { // Skip methods, only look at functions
			continue
		}

		// Check if it's a factory function (returns string and *Controller)
		prefix, controllerType := extractFactoryInfo(fn)
		if prefix != "" && controllerType != "" {
			info := &ControllerInfo{
				Prefix:   prefix,
				Type:     controllerType,
				FilePath: filePath,
				Methods:  []string{},
			}
			controllerTypes[controllerType] = info
			// Don't append yet - we need to find methods first
		}
	}

	// Second pass: Find methods for each controller type
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil { // Only look at methods
			continue
		}

		// Get the receiver type
		receiverType := getReceiverType(fn.Recv)
		
		// Check if this is a method on one of our controllers
		if info, exists := controllerTypes[receiverType]; exists {
			// Only include exported methods (start with uppercase)
			if fn.Name.IsExported() {
				info.Methods = append(info.Methods, fn.Name.Name)
			}
		}
	}

	// Build the final controller list with all methods
	for _, info := range controllerTypes {
		// Add embedded BaseController methods
		info.Methods = append(info.Methods, getBaseControllerMethods()...)
		info.Methods = removeDuplicates(info.Methods)
		controllers = append(controllers, *info)
	}

	return controllers, nil
}

// extractFactoryInfo extracts the prefix and controller type from a factory function
func extractFactoryInfo(fn *ast.FuncDecl) (string, string) {
	// Check if function returns (string, *Controller)
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 2 {
		return "", ""
	}

	// First result should be string
	first := fn.Type.Results.List[0]
	if ident, ok := first.Type.(*ast.Ident); !ok || ident.Name != "string" {
		return "", ""
	}

	// Second result should be *Controller
	second := fn.Type.Results.List[1]
	starExpr, ok := second.Type.(*ast.StarExpr)
	if !ok {
		return "", ""
	}

	// Get the controller type name
	controllerType := ""
	if ident, ok := starExpr.X.(*ast.Ident); ok {
		controllerType = ident.Name
	}

	// Now find the string literal in the return statement
	prefix := ""
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) >= 2 {
			if lit, ok := ret.Results[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				// Remove quotes from string literal
				prefix = strings.Trim(lit.Value, `"'`)
			}
		}
		return true
	})

	return prefix, controllerType
}

// getReceiverType extracts the receiver type from a method
func getReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	if field.Type == nil {
		return ""
	}

	// Handle pointer receivers
	if star, ok := field.Type.(*ast.StarExpr); ok {
		if ident, ok := star.X.(*ast.Ident); ok {
			return ident.Name
		}
	}

	// Handle value receivers
	if ident, ok := field.Type.(*ast.Ident); ok {
		return ident.Name
	}

	return ""
}

// getBaseControllerMethods returns common methods from embedded BaseController
func getBaseControllerMethods() []string {
	return []string{
		"Render",
		"Refresh",
		"Redirect",
		"RenderError",
		"RenderErrorMsg",
		"RenderString",
		"Use",
		"Atoi",
	}
}

// getSpecialControllers returns controllers from embedded packages like auth
func getSpecialControllers() []ControllerInfo {
	return []ControllerInfo{
		{
			Prefix:   "auth",
			Type:     "authentication.Controller",
			FilePath: "embedded:devtools/pkg/authentication",
			Methods: []string{
				"CurrentUser",
				"IsAuthenticated",
				"SigninURL",
				"SignupURL",
				"SignoutURL",
				"IsAdmin",
				"Username",
				"UserID",
			},
		},
	}
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(strs []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	
	return result
}