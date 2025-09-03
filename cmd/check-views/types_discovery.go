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

// TypeInfo represents any type that can be referenced in templates
type TypeInfo struct {
	Name       string                 // Type name (e.g., "ErrorData")
	Package    string                 // Package name (e.g., "templates")
	FilePath   string                 // Source file path
	Fields     map[string]FieldInfo   // Fields accessible in templates
	Methods    []MethodInfo           // Methods accessible in templates
	Source     string                 // "model", "internal", "controller-return"
	IsExported bool                   // Whether the type is exported
}

// DiscoverAllTypes discovers all types that could be used in templates
func DiscoverAllTypes(dir string) (map[string]*TypeInfo, error) {
	allTypes := make(map[string]*TypeInfo)
	
	// 1. Discover model types (already done)
	models, err := DiscoverModels(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover models: %w", err)
	}
	
	// Convert models to TypeInfo
	for name, model := range models {
		typeInfo := &TypeInfo{
			Name:       model.Name,
			Package:    model.Package,
			FilePath:   model.FilePath,
			Fields:     model.Fields,
			Methods:    model.Methods,
			Source:     "model",
			IsExported: true,
		}
		allTypes[name] = typeInfo
	}
	
	// 2. Discover internal types (templates, ai tools, etc.)
	internalTypes, err := DiscoverInternalTypes(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover internal types: %w", err)
	}
	
	for name, typeInfo := range internalTypes {
		allTypes[name] = typeInfo
	}
	
	// 3. Discover controller return types
	controllerTypes, err := DiscoverControllerReturnTypes(dir)
	if err != nil {
		// Don't fail completely, just warn
		if verbose {
			fmt.Printf("  ⚠️  Warning: Failed to discover controller return types: %v\n", err)
		}
	} else {
		for name, typeInfo := range controllerTypes {
			// Don't override if we already have this type from models or internal
			if _, exists := allTypes[name]; !exists {
				allTypes[name] = typeInfo
			}
		}
	}
	
	// 4. Add common interface{} types that are passed to templates
	addCommonTypes(allTypes)
	
	return allTypes, nil
}

// DiscoverInternalTypes discovers types in internal packages
func DiscoverInternalTypes(dir string) (map[string]*TypeInfo, error) {
	types := make(map[string]*TypeInfo)
	
	// Look for internal directory
	internalDir := filepath.Join(dir, "internal")
	
	// Walk through all Go files in internal directory
	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Internal directory might not exist
			if strings.Contains(err.Error(), "no such file") {
				return nil
			}
			return err
		}
		
		// Skip directories, non-Go files, and test files
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		// Parse the Go file for type definitions
		fileTypes, err := parseTypesFromFile(path, "internal")
		if err != nil {
			if verbose {
				fmt.Printf("  ⚠️  Error parsing internal file %s: %v\n", path, err)
			}
			return nil // Continue with other files
		}
		
		// Add types to map
		for name, typeInfo := range fileTypes {
			types[name] = typeInfo
		}
		
		return nil
	})
	
	if err != nil && !strings.Contains(err.Error(), "no such file") {
		return nil, fmt.Errorf("failed to walk internal directory: %w", err)
	}
	
	return types, nil
}

// parseTypesFromFile parses a single Go file for type definitions
func parseTypesFromFile(filePath string, source string) (map[string]*TypeInfo, error) {
	types := make(map[string]*TypeInfo)
	
	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	
	// Get package name
	packageName := node.Name.Name
	
	// Find all struct types
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			
			// Create TypeInfo
			typeInfo := &TypeInfo{
				Name:       typeSpec.Name.Name,
				Package:    packageName,
				FilePath:   filePath,
				Fields:     make(map[string]FieldInfo),
				Methods:    []MethodInfo{},
				Source:     source,
				IsExported: typeSpec.Name.IsExported(),
			}
			
			// Parse struct fields
			for _, field := range structType.Fields.List {
				fieldInfo := parseField(field)
				
				// Add each field name
				for _, name := range field.Names {
					if fieldInfo != nil {
						info := *fieldInfo
						info.Name = name.Name
						info.IsExported = name.IsExported()
						typeInfo.Fields[name.Name] = info
					}
				}
				
				// Handle embedded fields
				if len(field.Names) == 0 && fieldInfo != nil {
					typeName := getTypeName(field.Type)
					if typeName != "" {
						// Add as embedded field
						info := *fieldInfo
						info.Name = typeName
						info.IsExported = ast.IsExported(typeName)
						typeInfo.Fields[typeName] = info
					}
				}
			}
			
			types[typeSpec.Name.Name] = typeInfo
		}
	}
	
	// Find methods for these types
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			continue
		}
		
		// Get receiver type
		receiverType := getReceiverType(fn.Recv)
		
		// Check if this is a method on one of our types
		if typeInfo, exists := types[receiverType]; exists {
			// Add method if it's exported or matches template patterns
			if fn.Name.IsExported() || isTemplateAccessor(fn.Name.Name) {
				method := MethodInfo{
					Name:       fn.Name.Name,
					ReturnType: getReturnType(fn.Type),
				}
				typeInfo.Methods = append(typeInfo.Methods, method)
				
				// Also add as a field if it's a getter-style method
				if isGetterMethod(fn.Name.Name, fn.Type) {
					fieldName := fn.Name.Name
					typeInfo.Fields[fieldName] = FieldInfo{
						Name:       fieldName,
						Type:       method.ReturnType,
						IsExported: fn.Name.IsExported(),
					}
				}
			}
		}
	}
	
	return types, nil
}

// DiscoverControllerReturnTypes analyzes controller methods for return types
func DiscoverControllerReturnTypes(dir string) (map[string]*TypeInfo, error) {
	types := make(map[string]*TypeInfo)
	
	controllersDir := filepath.Join(dir, "controllers")
	
	// Walk through controller files
	err := filepath.WalkDir(controllersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		// Parse file and look for Render calls to find types
		if err := findRenderTypes(path, types); err != nil {
			if verbose {
				fmt.Printf("  ⚠️  Error analyzing controller %s: %v\n", path, err)
			}
		}
		
		return nil
	})
	
	return types, err
}

// findRenderTypes looks for c.Render() calls and extracts the data types
func findRenderTypes(filePath string, types map[string]*TypeInfo) error {
	// This is a simplified version - in practice, we'd need more sophisticated analysis
	// using go/types package for full type resolution
	
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	
	// Look for Render calls and extract the data parameter type
	// This would need enhancement with go/types for proper type resolution
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		
		// Look for Render method calls
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Render" && len(call.Args) >= 3 {
				// The third argument is typically the data
				// Here we'd need go/types to resolve the actual type
				// For now, just mark that we found a Render call
				if verbose {
					// Could enhance this to extract actual types
				}
			}
		}
		
		return true
	})
	
	return nil
}

// isTemplateAccessor checks if a method name looks like a template accessor
func isTemplateAccessor(name string) bool {
	// Methods like html_url(), stargazers_count() for template compatibility
	return strings.Contains(name, "_") && !strings.HasPrefix(name, "_")
}

// isGetterMethod checks if this looks like a getter method
func isGetterMethod(name string, fnType *ast.FuncType) bool {
	if fnType == nil || fnType.Results == nil {
		return false
	}
	
	// Simple getter: no parameters, returns one value
	hasNoParams := fnType.Params == nil || len(fnType.Params.List) == 0
	hasOneReturn := len(fnType.Results.List) == 1
	
	return hasNoParams && hasOneReturn
}

// getReturnType extracts a simplified return type from a function
func getReturnType(fnType *ast.FuncType) string {
	if fnType == nil || fnType.Results == nil || len(fnType.Results.List) == 0 {
		return ""
	}
	
	// Get the first return type
	firstResult := fnType.Results.List[0]
	return getTypeName(firstResult.Type)
}

// addCommonTypes adds common types that are often passed to templates
func addCommonTypes(types map[string]*TypeInfo) {
	// Add error type (for error messages)
	if _, exists := types["error"]; !exists {
		types["error"] = &TypeInfo{
			Name:    "error",
			Package: "builtin",
			Source:  "builtin",
			Fields: map[string]FieldInfo{
				"Error": {
					Name:       "Error",
					Type:       "string",
					IsExported: true,
				},
			},
			Methods: []MethodInfo{
				{
					Name:       "Error",
					ReturnType: "string",
				},
			},
		}
	}
}

// Helper to check if a type should be included for template validation
func shouldIncludeType(typeInfo *TypeInfo) bool {
	// Include all exported types and types from templates package
	return typeInfo.IsExported || typeInfo.Package == "templates"
}