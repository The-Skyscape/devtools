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

// ModelInfo represents a discovered model struct
type ModelInfo struct {
	Name     string               // e.g., "Workspace"
	Package  string               // e.g., "models"
	FilePath string               // e.g., "models/workspace.go"
	Fields   map[string]FieldInfo // Field name -> type info
	Methods  []MethodInfo         // Methods that can be called on this model
}

// FieldInfo represents a struct field
type FieldInfo struct {
	Name       string // Field name
	Type       string // Type name (e.g., "string", "int", "User")
	IsPtr      bool   // Whether it's a pointer type
	IsSlice    bool   // Whether it's a slice
	IsExported bool   // Whether the field is exported
}

// MethodInfo represents a method on a model
type MethodInfo struct {
	Name         string // Method name
	ReturnType   string // Return type (simplified)
	ReturnsError bool   // Whether it returns an error as second value
}

// DiscoverModels finds all model structs in the models directory
func DiscoverModels(dir string) (map[string]*ModelInfo, error) {
	models := make(map[string]*ModelInfo)

	// Find models directory
	modelsDir := filepath.Join(dir, "models")

	// Walk through all Go files in models directory
	err := filepath.WalkDir(modelsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-Go files, and test files
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse the Go file
		fileModels, err := parseModelFile(path)
		if err != nil {
			if verbose {
				fmt.Printf("  ⚠️  Error parsing model file %s: %v\n", path, err)
			}
			return nil // Continue with other files
		}

		// Add models to map
		for name, model := range fileModels {
			models[name] = model
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk models directory: %w", err)
	}

	// Add base model fields that are commonly embedded
	addBaseModelFields(models)

	return models, nil
}

// parseModelFile parses a single Go file for model definitions
func parseModelFile(filePath string) (map[string]*ModelInfo, error) {
	models := make(map[string]*ModelInfo)

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Get package name
	packageName := node.Name.Name

	// First pass: Find all struct types
	structTypes := make(map[string]*ast.StructType)
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

			// Only include exported types
			if typeSpec.Name.IsExported() {
				structTypes[typeSpec.Name.Name] = structType
			}
		}
	}

	// Second pass: Extract fields from each struct
	for typeName, structType := range structTypes {
		model := &ModelInfo{
			Name:     typeName,
			Package:  packageName,
			FilePath: filePath,
			Fields:   make(map[string]FieldInfo),
			Methods:  []MethodInfo{},
		}

		// Parse struct fields
		for _, field := range structType.Fields.List {
			fieldInfo := parseField(field)

			// Add each field name (a field can have multiple names)
			for _, name := range field.Names {
				if fieldInfo != nil {
					info := *fieldInfo
					info.Name = name.Name
					info.IsExported = name.IsExported()
					model.Fields[name.Name] = info
				}
			}

			// Handle embedded fields (no names)
			if len(field.Names) == 0 && fieldInfo != nil {
				// Embedded field - extract type name
				typeName := getTypeName(field.Type)
				if typeName != "" {
					// Mark as embedded by prefixing with "*"
					model.Fields["*"+typeName] = *fieldInfo
				}
			}
		}

		models[typeName] = model
	}

	// Third pass: Find methods for each struct
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			continue
		}

		// Get receiver type
		receiverType := getReceiverType(fn.Recv)

		// Check if this method belongs to one of our models
		if model, exists := models[receiverType]; exists {
			// Only include exported methods
			if fn.Name.IsExported() {
				method := parseMethod(fn)
				if method != nil {
					model.Methods = append(model.Methods, *method)
				}
			}
		}
	}

	return models, nil
}

// parseField extracts field information from an AST field
func parseField(field *ast.Field) *FieldInfo {
	if field.Type == nil {
		return nil
	}

	info := &FieldInfo{}

	// Parse the type
	switch t := field.Type.(type) {
	case *ast.Ident:
		// Simple type like string, int, User
		info.Type = t.Name

	case *ast.StarExpr:
		// Pointer type like *User
		info.IsPtr = true
		if ident, ok := t.X.(*ast.Ident); ok {
			info.Type = ident.Name
		} else if sel, ok := t.X.(*ast.SelectorExpr); ok {
			// Qualified type like *time.Time
			if pkg, ok := sel.X.(*ast.Ident); ok {
				info.Type = pkg.Name + "." + sel.Sel.Name
			}
		}

	case *ast.ArrayType:
		// Slice type like []string
		info.IsSlice = true
		if ident, ok := t.Elt.(*ast.Ident); ok {
			info.Type = ident.Name
		} else if star, ok := t.Elt.(*ast.StarExpr); ok {
			// Slice of pointers like []*User
			info.IsPtr = true
			if ident, ok := star.X.(*ast.Ident); ok {
				info.Type = ident.Name
			}
		}

	case *ast.SelectorExpr:
		// Qualified type like time.Time
		if pkg, ok := t.X.(*ast.Ident); ok {
			info.Type = pkg.Name + "." + t.Sel.Name
		}

	default:
		// Complex type - just get string representation
		info.Type = fmt.Sprintf("%T", t)
	}

	return info
}

// parseMethod extracts method information from a function declaration
func parseMethod(fn *ast.FuncDecl) *MethodInfo {
	if fn.Type == nil || fn.Type.Results == nil {
		return nil
	}

	method := &MethodInfo{
		Name: fn.Name.Name,
	}

	// Parse return types
	results := fn.Type.Results.List
	if len(results) > 0 {
		// Get the first return type
		firstResult := results[0]
		method.ReturnType = getTypeName(firstResult.Type)

		// Check if it returns an error as second value
		if len(results) > 1 {
			secondResult := results[1]
			if errorType := getTypeName(secondResult.Type); errorType == "error" {
				method.ReturnsError = true
			}
		}
	}

	return method
}

// getTypeName extracts a simple type name from an AST expression
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return getTypeName(t.X)
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return pkg.Name + "." + t.Sel.Name
		}
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	}
	return ""
}

// addBaseModelFields adds commonly embedded model fields
func addBaseModelFields(models map[string]*ModelInfo) {
	// Common embedded fields from application.Model
	baseFields := map[string]FieldInfo{
		"ID": {
			Name:       "ID",
			Type:       "string",
			IsExported: true,
		},
		"CreatedAt": {
			Name:       "CreatedAt",
			Type:       "time.Time",
			IsExported: true,
		},
		"UpdatedAt": {
			Name:       "UpdatedAt",
			Type:       "time.Time",
			IsExported: true,
		},
	}

	// Add base fields to models that embed application.Model
	for _, model := range models {
		// Check if model embeds application.Model
		if _, hasEmbedded := model.Fields["*application.Model"]; hasEmbedded {
			// Add base fields if not already present
			for name, field := range baseFields {
				if _, exists := model.Fields[name]; !exists {
					model.Fields[name] = field
				}
			}
		}
	}
}

// GetModelFields returns all field names for a given model type
func GetModelFields(models map[string]*ModelInfo, modelType string) []string {
	model, exists := models[modelType]
	if !exists {
		return nil
	}

	var fields []string
	for name, field := range model.Fields {
		// Skip embedded type markers
		if strings.HasPrefix(name, "*") {
			continue
		}
		// Only include exported fields
		if field.IsExported {
			fields = append(fields, name)
		}
	}

	return fields
}

// GetModelMethods returns all method names for a given model type
func GetModelMethods(models map[string]*ModelInfo, modelType string) []string {
	model, exists := models[modelType]
	if !exists {
		return nil
	}

	var methods []string
	for _, method := range model.Methods {
		methods = append(methods, method.Name)
	}

	return methods
}

// ResolveFieldType returns the type of a field access chain
func ResolveFieldType(models map[string]*ModelInfo, baseType string, fieldChain []string) (string, bool) {
	currentType := baseType

	for _, fieldName := range fieldChain {
		model, exists := models[currentType]
		if !exists {
			return "", false
		}

		// Check if it's a field
		if field, hasField := model.Fields[fieldName]; hasField {
			currentType = field.Type
			// Remove package qualifiers for local types
			if strings.Contains(currentType, ".") {
				parts := strings.Split(currentType, ".")
				currentType = parts[len(parts)-1]
			}
			continue
		}

		// Check if it's a method
		for _, method := range model.Methods {
			if method.Name == fieldName {
				currentType = method.ReturnType
				// Remove pointer and package qualifiers
				currentType = strings.TrimPrefix(currentType, "*")
				if strings.Contains(currentType, ".") {
					parts := strings.Split(currentType, ".")
					currentType = parts[len(parts)-1]
				}
				break
			}
		}
	}

	return currentType, true
}
