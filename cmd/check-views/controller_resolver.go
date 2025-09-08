package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ControllerResolver uses go/packages to fully resolve controller methods including embedded types
type ControllerResolver struct {
	pkg         *packages.Package
	controllers map[string]*ControllerInfo
}

// NewControllerResolver creates a resolver that understands embedded types
func NewControllerResolver(dir string) (*ControllerResolver, error) {
	controllersDir := filepath.Join(dir, "controllers")

	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax |
			packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedDeps,
		Dir: controllersDir,
	}

	// Load the controllers package
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load controllers package: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found in controllers directory")
	}

	pkg := pkgs[0]

	// Check for errors
	if len(pkg.Errors) > 0 {
		// Log but don't fail
		if verbose {
			log.Printf("Package has errors: %v", pkg.Errors)
		}
	}

	return &ControllerResolver{
		pkg:         pkg,
		controllers: make(map[string]*ControllerInfo),
	}, nil
}

// DiscoverControllers finds all controllers and their methods including embedded ones
func (cr *ControllerResolver) DiscoverControllers() ([]ControllerInfo, error) {
	// First find factory functions to identify controllers
	if err := cr.findFactoryFunctions(); err != nil {
		return nil, fmt.Errorf("failed to find factory functions: %w", err)
	}

	// Then resolve all methods including embedded ones
	if err := cr.resolveControllerMethods(); err != nil {
		return nil, fmt.Errorf("failed to resolve methods: %w", err)
	}

	// Convert map to slice
	var result []ControllerInfo
	for _, ctrl := range cr.controllers {
		result = append(result, *ctrl)
	}

	return result, nil
}

// findFactoryFunctions identifies controller factory functions
func (cr *ControllerResolver) findFactoryFunctions() error {
	for _, file := range cr.pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil { // Only look at functions, not methods
				continue
			}

			// Check if it's a factory function
			prefix, controllerType := cr.extractFactoryInfo(fn)
			if prefix != "" && controllerType != "" {
				// Get the file path
				filePath := ""
				for i, f := range cr.pkg.Syntax {
					if f == file {
						if i < len(cr.pkg.CompiledGoFiles) {
							filePath = cr.pkg.CompiledGoFiles[i]
						} else if len(cr.pkg.GoFiles) > i {
							filePath = cr.pkg.GoFiles[i]
						}
						break
					}
				}

				cr.controllers[controllerType] = &ControllerInfo{
					Prefix:        prefix,
					Type:          controllerType,
					FilePath:      filePath,
					Methods:       []string{},
					MethodDetails: make(map[string]*MethodDetail),
				}
			}
		}
	}

	return nil
}

// extractFactoryInfo extracts controller info from a factory function
func (cr *ControllerResolver) extractFactoryInfo(fn *ast.FuncDecl) (string, string) {
	// Check return type (should return string and *Controller)
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 2 {
		return "", ""
	}

	// First result should be string (the prefix)
	firstResult := fn.Type.Results.List[0]
	if ident, ok := firstResult.Type.(*ast.Ident); !ok || ident.Name != "string" {
		return "", ""
	}

	// Second result should be a pointer to a controller
	secondResult := fn.Type.Results.List[1]
	var controllerType string

	switch t := secondResult.Type.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			controllerType = ident.Name
		}
	}

	if controllerType == "" || !strings.HasSuffix(controllerType, "Controller") {
		return "", ""
	}

	// Parse function body to find the returned string literal
	var prefix string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) >= 1 {
			if lit, ok := ret.Results[0].(*ast.BasicLit); ok {
				prefix = strings.Trim(lit.Value, `"`)
				return false
			}
		}
		return true
	})

	return prefix, controllerType
}

// resolveControllerMethods finds all methods including from embedded types
func (cr *ControllerResolver) resolveControllerMethods() error {
	if cr.pkg.Types == nil {
		return fmt.Errorf("package types not available")
	}

	scope := cr.pkg.Types.Scope()

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)

		// Check if it's a type we care about (a controller)
		typeName, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}

		// Check if this is one of our controllers
		ctrl, exists := cr.controllers[typeName.Name()]
		if !exists {
			continue
		}

		// Get the named type
		named, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}

		// Collect all methods including embedded ones with details
		methods, methodDetails := cr.collectAllMethodsWithDetails(named)
		ctrl.Methods = methods
		ctrl.MethodDetails = methodDetails

		if verbose {
			log.Printf("Controller %s (%s) has %d methods including embedded",
				ctrl.Prefix, ctrl.Type, len(methods))
		}
	}

	return nil
}

// collectAllMethodsWithDetails collects all methods with return type information
func (cr *ControllerResolver) collectAllMethodsWithDetails(named *types.Named) ([]string, map[string]*MethodDetail) {
	methodSet := make(map[string]*MethodDetail)

	// Get direct methods
	for i := 0; i < named.NumMethods(); i++ {
		method := named.Method(i)
		if method.Exported() {
			detail := cr.extractMethodDetail(method)
			if detail != nil {
				methodSet[method.Name()] = detail
			}
		}
	}

	// Get methods from embedded types
	if structType, ok := named.Underlying().(*types.Struct); ok {
		cr.collectEmbeddedMethodsWithDetails(structType, methodSet)
	}

	// Convert to slice for backward compatibility
	var methods []string
	for name := range methodSet {
		methods = append(methods, name)
	}

	return methods, methodSet
}

// extractMethodDetail extracts detailed information from a method
func (cr *ControllerResolver) extractMethodDetail(method *types.Func) *MethodDetail {
	sig, ok := method.Type().(*types.Signature)
	if !ok {
		return nil
	}

	detail := &MethodDetail{
		Name:       method.Name(),
		IsExported: method.Exported(),
		Returns:    []string{},
	}

	// Extract return types
	if results := sig.Results(); results != nil {
		for i := 0; i < results.Len(); i++ {
			param := results.At(i)
			typeName := cr.getTypeName(param.Type())
			if typeName != "" {
				detail.Returns = append(detail.Returns, typeName)
			}
		}
	}

	return detail
}

// getTypeName extracts a readable type name from a types.Type
func (cr *ControllerResolver) getTypeName(t types.Type) string {
	switch typ := t.(type) {
	case *types.Named:
		return typ.Obj().Name()
	case *types.Pointer:
		return "*" + cr.getTypeName(typ.Elem())
	case *types.Slice:
		return "[]" + cr.getTypeName(typ.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", typ.Len(), cr.getTypeName(typ.Elem()))
	case *types.Basic:
		return typ.Name()
	case *types.Interface:
		if typ.Empty() {
			return "interface{}"
		}
		return "interface"
	case *types.Struct:
		return "struct"
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", cr.getTypeName(typ.Key()), cr.getTypeName(typ.Elem()))
	default:
		return ""
	}
}

// collectAllMethods collects all methods including from embedded types (legacy)
func (cr *ControllerResolver) collectAllMethods(named *types.Named) []string {
	methods, _ := cr.collectAllMethodsWithDetails(named)
	return methods
}

// collectEmbeddedMethodsWithDetails recursively collects methods with details from embedded types
func (cr *ControllerResolver) collectEmbeddedMethodsWithDetails(structType *types.Struct, methodSet map[string]*MethodDetail) {
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		// Check if it's an embedded field
		if !field.Embedded() {
			continue
		}

		// Get the type of the embedded field
		fieldType := field.Type()

		// Handle pointer types
		if ptr, ok := fieldType.(*types.Pointer); ok {
			fieldType = ptr.Elem()
		}

		// If it's a named type, get its methods
		if namedType, ok := fieldType.(*types.Named); ok {
			// Get methods from this embedded type
			for j := 0; j < namedType.NumMethods(); j++ {
				method := namedType.Method(j)
				if method.Exported() {
					// Don't override if already present (shadowing)
					if _, exists := methodSet[method.Name()]; !exists {
						detail := cr.extractMethodDetail(method)
						if detail != nil {
							methodSet[method.Name()] = detail
						}
					}
				}
			}

			// Recursively check embedded types within this type
			if innerStruct, ok := namedType.Underlying().(*types.Struct); ok {
				cr.collectEmbeddedMethodsWithDetails(innerStruct, methodSet)
			}
		}
	}
}

// collectEmbeddedMethods recursively collects methods from embedded types (legacy)
func (cr *ControllerResolver) collectEmbeddedMethods(structType *types.Struct, methodSet map[string]bool) {
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		// Check if it's an embedded field
		if !field.Embedded() {
			continue
		}

		// Get the type of the embedded field
		fieldType := field.Type()

		// Handle pointer types
		if ptr, ok := fieldType.(*types.Pointer); ok {
			fieldType = ptr.Elem()
		}

		// If it's a named type, get its methods
		if named, ok := fieldType.(*types.Named); ok {
			// Get methods of the embedded type
			for i := 0; i < named.NumMethods(); i++ {
				method := named.Method(i)
				if method.Exported() {
					methodSet[method.Name()] = true
				}
			}

			// Recursively check for embedded types in the embedded type
			if embeddedStruct, ok := named.Underlying().(*types.Struct); ok {
				cr.collectEmbeddedMethods(embeddedStruct, methodSet)
			}
		}
	}
}

// DiscoverControllersEnhanced uses the enhanced resolver if possible, falls back to basic discovery
func DiscoverControllersEnhanced(dir string) ([]ControllerInfo, error) {
	// Try enhanced resolution first
	resolver, err := NewControllerResolver(dir)
	if err != nil {
		if verbose {
			log.Printf("Failed to use enhanced controller resolution, falling back: %v", err)
		}
		// Fall back to basic discovery
		return DiscoverControllers(dir)
	}

	controllers, err := resolver.DiscoverControllers()
	if err != nil {
		if verbose {
			log.Printf("Enhanced resolution failed, falling back: %v", err)
		}
		// Fall back to basic discovery
		return DiscoverControllers(dir)
	}

	// If we got no controllers, fall back
	if len(controllers) == 0 {
		if verbose {
			log.Printf("No controllers found with enhanced resolution, falling back")
		}
		return DiscoverControllers(dir)
	}

	return controllers, nil
}
