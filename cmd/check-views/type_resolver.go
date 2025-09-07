package main

import (
	"fmt"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

// TypeResolver uses go/packages for full type resolution
type TypeResolver struct {
	packages map[string]*packages.Package
	typeMap  map[string]*ResolvedType
}

// ResolvedType represents a fully resolved type with all its fields and methods
type ResolvedType struct {
	Name       string
	Package    string
	Fields     map[string]*types.Var
	Methods    map[string]*types.Func
	Underlying types.Type
}

// NewTypeResolver creates a new type resolver for the given directory
func NewTypeResolver(dir string) (*TypeResolver, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax |
			packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports,
		Dir: dir,
	}

	// Load all packages in the directory
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Check for errors in loaded packages
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			// Log but don't fail - some packages may have errors
			if verbose {
				log.Printf("Package %s has errors: %v", pkg.PkgPath, pkg.Errors)
			}
		}
	}

	tr := &TypeResolver{
		packages: make(map[string]*packages.Package),
		typeMap:  make(map[string]*ResolvedType),
	}

	// Index packages by path
	for _, pkg := range pkgs {
		tr.packages[pkg.PkgPath] = pkg

		// Build type map for this package
		tr.buildTypeMap(pkg)
	}

	// Add common generic types that templates use
	tr.addCommonGenericTypes()

	return tr, nil
}

// buildTypeMap builds a map of all types in a package
func (tr *TypeResolver) buildTypeMap(pkg *packages.Package) {
	if pkg.Types == nil {
		return
	}

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)

		// We're interested in type names
		if typeName, ok := obj.(*types.TypeName); ok {
			// Skip built-in types
			if typeName.Pkg() == nil {
				continue
			}

			// Get the underlying type
			namedType, ok := typeName.Type().(*types.Named)
			if !ok {
				continue
			}

			// Create resolved type
			rt := &ResolvedType{
				Name:       typeName.Name(),
				Package:    typeName.Pkg().Path(),
				Fields:     make(map[string]*types.Var),
				Methods:    make(map[string]*types.Func),
				Underlying: namedType.Underlying(),
			}

			// Get struct fields if it's a struct
			if structType, ok := namedType.Underlying().(*types.Struct); ok {
				for i := 0; i < structType.NumFields(); i++ {
					field := structType.Field(i)
					rt.Fields[field.Name()] = field
				}
			}

			// Get methods
			for i := 0; i < namedType.NumMethods(); i++ {
				method := namedType.Method(i)
				rt.Methods[method.Name()] = method
			}

			// Store in map with fully qualified name
			fullName := fmt.Sprintf("%s.%s", typeName.Pkg().Path(), typeName.Name())
			tr.typeMap[fullName] = rt

			// Also store with just the type name for easier lookup
			tr.typeMap[typeName.Name()] = rt
		}
	}
}

// GetType returns a resolved type by name
func (tr *TypeResolver) GetType(name string) *ResolvedType {
	// Try exact match first
	if rt, ok := tr.typeMap[name]; ok {
		return rt
	}

	// Try with different package prefixes
	for fullName, rt := range tr.typeMap {
		if strings.HasSuffix(fullName, "."+name) {
			return rt
		}
	}

	return nil
}

// GetControllerReturnType analyzes a controller method to find what type it returns
func (tr *TypeResolver) GetControllerReturnType(controllerName, methodName string) *ResolvedType {
	// Look for the controller type
	controllerType := tr.GetType(controllerName + "Controller")
	if controllerType == nil {
		return nil
	}

	// Look for the method
	if method, ok := controllerType.Methods[methodName]; ok {
		sig, ok := method.Type().(*types.Signature)
		if !ok {
			return nil
		}

		// Get return type
		results := sig.Results()
		if results.Len() == 0 {
			return nil
		}

		// Get the first return value (usually the data for templates)
		returnType := results.At(0).Type()

		// Resolve the type
		return tr.resolveType(returnType)
	}

	return nil
}

// resolveType converts a types.Type to our ResolvedType
func (tr *TypeResolver) resolveType(t types.Type) *ResolvedType {
	switch typ := t.(type) {
	case *types.Named:
		// Look up the named type
		if typ.Obj() != nil && typ.Obj().Pkg() != nil {
			fullName := fmt.Sprintf("%s.%s", typ.Obj().Pkg().Path(), typ.Obj().Name())
			if rt, ok := tr.typeMap[fullName]; ok {
				return rt
			}
		}

	case *types.Pointer:
		// Dereference pointer and try again
		return tr.resolveType(typ.Elem())

	case *types.Slice:
		// For slices, we might want to know the element type
		return tr.resolveType(typ.Elem())
	}

	return nil
}

// GetFieldType returns the type of a field in a resolved type
func (rt *ResolvedType) GetFieldType(fieldName string) types.Type {
	if field, ok := rt.Fields[fieldName]; ok {
		return field.Type()
	}

	// Check if it's a method that acts like a field (getter)
	if method, ok := rt.Methods[fieldName]; ok {
		sig, ok := method.Type().(*types.Signature)
		if ok && sig.Params().Len() == 0 && sig.Results().Len() == 1 {
			return sig.Results().At(0).Type()
		}
	}

	return nil
}

// HasField checks if a type has a field or method with the given name
func (rt *ResolvedType) HasField(fieldName string) bool {
	if _, ok := rt.Fields[fieldName]; ok {
		return true
	}

	if _, ok := rt.Methods[fieldName]; ok {
		return true
	}

	return false
}

// addCommonGenericTypes adds common types that templates frequently use
func (tr *TypeResolver) addCommonGenericTypes() {
	// Generic error type
	tr.typeMap["GenericError"] = &ResolvedType{
		Name:    "GenericError",
		Package: "generic",
		Fields: map[string]*types.Var{
			"Error": nil, // Template uses .Error
		},
		Methods:    make(map[string]*types.Func),
		Underlying: nil,
	}

	// Generic success type
	tr.typeMap["GenericSuccess"] = &ResolvedType{
		Name:    "GenericSuccess",
		Package: "generic",
		Fields: map[string]*types.Var{
			"Success": nil, // Template uses .Success
			"Message": nil, // Template uses .Message
		},
		Methods:    make(map[string]*types.Func),
		Underlying: nil,
	}

	// Generic content type (for streaming, etc)
	tr.typeMap["GenericContent"] = &ResolvedType{
		Name:    "GenericContent",
		Package: "generic",
		Fields: map[string]*types.Var{
			"Content": nil, // Template uses .Content
		},
		Methods:    make(map[string]*types.Func),
		Underlying: nil,
	}

	// Generic pagination type
	tr.typeMap["GenericPagination"] = &ResolvedType{
		Name:    "GenericPagination",
		Package: "generic",
		Fields: map[string]*types.Var{
			"HasMore":  nil, // Template uses .HasMore
			"NextPage": nil, // Template uses .NextPage
			"Endpoint": nil, // Template uses .Endpoint
		},
		Methods:    make(map[string]*types.Func),
		Underlying: nil,
	}

	// GitHub API types (for import functionality)
	tr.typeMap["GitHubRepo"] = &ResolvedType{
		Name:    "GitHubRepo",
		Package: "github",
		Fields: map[string]*types.Var{
			"name":             nil, // GitHub API uses lowercase
			"description":      nil,
			"html_url":         nil,
			"private":          nil,
			"language":         nil,
			"stargazers_count": nil,
		},
		Methods:    make(map[string]*types.Func),
		Underlying: nil,
	}
}

// GetTypeInfo converts ResolvedType to TypeInfo for compatibility with existing code
func (rt *ResolvedType) GetTypeInfo() *TypeInfo {
	ti := &TypeInfo{
		Name:       rt.Name,
		Package:    rt.Package,
		Fields:     make(map[string]FieldInfo),
		Methods:    []MethodInfo{},
		Source:     "resolved",
		IsExported: true,
	}

	// Convert fields
	for name, field := range rt.Fields {
		ti.Fields[name] = FieldInfo{
			Name:       name,
			Type:       field.Type().String(),
			IsExported: field.Exported(),
		}
	}

	// Convert methods
	for name, method := range rt.Methods {
		if !method.Exported() {
			continue
		}

		returnType := ""
		if sig, ok := method.Type().(*types.Signature); ok && sig.Results().Len() > 0 {
			returnType = sig.Results().At(0).Type().String()
		}

		ti.Methods = append(ti.Methods, MethodInfo{
			Name:       name,
			ReturnType: returnType,
		})
	}

	return ti
}

// AnalyzeTemplateContext analyzes what type is passed to a template
func (tr *TypeResolver) AnalyzeTemplateContext(pkg *packages.Package, templateName string) *ResolvedType {
	// This would analyze c.Render() calls to determine what type is passed
	// to each template. For now, we'll implement a simplified version.

	// Look for common patterns:
	// 1. If it's a repo-* template, it likely gets a Repository type
	// 2. If it's a settings-* template, it likely gets settings data
	// 3. Partials often get specific types or the parent context

	if strings.HasPrefix(templateName, "repo-") {
		return tr.GetType("Repository")
	}

	if strings.HasPrefix(templateName, "settings-") {
		return tr.GetType("Settings")
	}

	// Default: could be anything
	return nil
}
