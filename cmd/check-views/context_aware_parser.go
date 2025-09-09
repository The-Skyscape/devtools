package main

import (
	"fmt"
	"log"
	"strings"
	"text/template/parse"
)

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// TemplateScope represents the current parsing context with type information
type TemplateScope struct {
	Parent    *TemplateScope
	Variables map[string]*ResolvedType // $var -> Type mapping
	DotType   *ResolvedType             // Current {{.}} context type
}

// NewTemplateScope creates a new scope, optionally inheriting from parent
func NewTemplateScope(parent *TemplateScope) *TemplateScope {
	scope := &TemplateScope{
		Parent:    parent,
		Variables: make(map[string]*ResolvedType),
	}
	
	// Inherit parent's variables (but can be shadowed)
	if parent != nil {
		for k, v := range parent.Variables {
			scope.Variables[k] = v
		}
		// Don't inherit DotType - it's context specific
	}
	
	return scope
}

// LookupVariable finds a variable in current or parent scopes
func (s *TemplateScope) LookupVariable(name string) *ResolvedType {
	if typ, ok := s.Variables[name]; ok {
		return typ
	}
	if s.Parent != nil {
		return s.Parent.LookupVariable(name)
	}
	return nil
}

// ContextAwareParser parses templates with full type tracking
type ContextAwareParser struct {
	controllers map[string]*ControllerInfo
	resolver    *TypeResolver
	verbose     bool
}

// NewContextAwareParser creates a parser with type awareness
func NewContextAwareParser(controllers []ControllerInfo, resolver *TypeResolver, verbose bool) *ContextAwareParser {
	controllerMap := make(map[string]*ControllerInfo)
	for i := range controllers {
		controllerMap[controllers[i].Prefix] = &controllers[i]
	}
	
	return &ContextAwareParser{
		controllers: controllerMap,
		resolver:    resolver,
		verbose:     verbose,
	}
}

// ParseTemplate parses a template with context tracking
func (p *ContextAwareParser) ParseTemplate(tree *parse.Tree, filename string) ([]TemplateReference, []FieldReference) {
	var templateRefs []TemplateReference
	var fieldRefs []FieldReference
	
	if tree == nil || tree.Root == nil {
		return templateRefs, fieldRefs
	}
	
	// Start with empty scope
	scope := NewTemplateScope(nil)
	
	// Process the tree
	p.walkNode(tree.Root, filename, scope, &templateRefs, &fieldRefs)
	
	return templateRefs, fieldRefs
}

// walkNode recursively processes template nodes with context
func (p *ContextAwareParser) walkNode(node parse.Node, file string, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if node == nil {
		return
	}
	
	switch n := node.(type) {
	case *parse.ActionNode:
		// {{expression}}
		if n.Pipe != nil {
			p.processPipe(n.Pipe, file, n.Line, scope, templateRefs, fieldRefs)
		}
		
	case *parse.IfNode:
		// {{if pipeline}} T1 {{else}} T2 {{end}}
		if n.Pipe != nil {
			// Check for assignment: {{if $var := expr}}
			newScope := p.processPipeWithAssignment(n.Pipe, file, n.Line, scope, templateRefs, fieldRefs)
			if n.List != nil {
				p.walkNode(n.List, file, newScope, templateRefs, fieldRefs)
			}
			if n.ElseList != nil {
				// Else branch doesn't have the assignment
				p.walkNode(n.ElseList, file, scope, templateRefs, fieldRefs)
			}
		}
		
	case *parse.RangeNode:
		// {{range pipeline}} T1 {{else}} T2 {{end}}
		if n.Pipe != nil {
			elementType := p.resolveRangeType(n.Pipe, file, n.Line, scope, templateRefs, fieldRefs)
			if elementType != nil && n.List != nil {
				// Create new scope with . bound to element type
				rangeScope := NewTemplateScope(scope)
				rangeScope.DotType = elementType
				
				// Check for index/value variables: {{range $i, $v := expr}}
				if len(n.Pipe.Decl) > 0 {
					// Handle variable declarations
					if len(n.Pipe.Decl) == 1 {
						// {{range $v := expr}}
						rangeScope.Variables[n.Pipe.Decl[0].Ident[0]] = elementType
					} else if len(n.Pipe.Decl) == 2 {
						// {{range $i, $v := expr}}
						// First is index (int for slices, key type for maps)
						rangeScope.Variables[n.Pipe.Decl[0].Ident[0]] = p.resolver.GetType("int")
						rangeScope.Variables[n.Pipe.Decl[1].Ident[0]] = elementType
					}
				}
				
				p.walkNode(n.List, file, rangeScope, templateRefs, fieldRefs)
			}
			if n.ElseList != nil {
				p.walkNode(n.ElseList, file, scope, templateRefs, fieldRefs)
			}
		}
		
	case *parse.WithNode:
		// {{with pipeline}} T1 {{else}} T2 {{end}}
		if n.Pipe != nil {
			withType := p.resolvePipeType(n.Pipe, file, n.Line, scope, templateRefs, fieldRefs)
			if withType != nil && n.List != nil {
				// Create new scope with . bound to with type
				withScope := NewTemplateScope(scope)
				withScope.DotType = withType
				
				// Check for variable assignment: {{with $var := expr}}
				if len(n.Pipe.Decl) > 0 && len(n.Pipe.Decl[0].Ident) > 0 {
					withScope.Variables[n.Pipe.Decl[0].Ident[0]] = withType
				}
				
				p.walkNode(n.List, file, withScope, templateRefs, fieldRefs)
			}
			if n.ElseList != nil {
				p.walkNode(n.ElseList, file, scope, templateRefs, fieldRefs)
			}
		}
		
	case *parse.ListNode:
		// Sequence of nodes
		if n != nil && n.Nodes != nil {
			for _, child := range n.Nodes {
				p.walkNode(child, file, scope, templateRefs, fieldRefs)
			}
		}
	}
}

// processPipe processes a pipeline and tracks any assignments
func (p *ContextAwareParser) processPipe(pipe *parse.PipeNode, file string, line int, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if pipe == nil {
		return
	}
	
	// Check for variable assignment
	if len(pipe.Decl) > 0 {
		// This is an assignment like {{$var := expr}}
		if len(pipe.Cmds) > 0 {
			typ := p.resolveCommandType(pipe.Cmds[0], file, line, scope, templateRefs, fieldRefs)
			if typ != nil && len(pipe.Decl[0].Ident) > 0 {
				varName := pipe.Decl[0].Ident[0]
				scope.Variables[varName] = typ
				if p.verbose {
					log.Printf("Assigned variable %s with type %s in %s:%d", varName, typ.Name, file, line)
				}
			}
		}
	} else {
		// Regular pipeline without assignment
		for _, cmd := range pipe.Cmds {
			p.processCommand(cmd, file, line, scope, templateRefs, fieldRefs)
		}
	}
}

// processPipeWithAssignment processes a pipe that might have assignment and returns appropriate scope
func (p *ContextAwareParser) processPipeWithAssignment(pipe *parse.PipeNode, file string, line int, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) *TemplateScope {
	if pipe == nil {
		return scope
	}
	
	// Check for assignment in if: {{if $var := expr}}
	if len(pipe.Decl) > 0 && len(pipe.Cmds) > 0 {
		newScope := NewTemplateScope(scope)
		typ := p.resolveCommandType(pipe.Cmds[0], file, line, scope, templateRefs, fieldRefs)
		if typ != nil && len(pipe.Decl[0].Ident) > 0 {
			varName := pipe.Decl[0].Ident[0]
			newScope.Variables[varName] = typ
		}
		return newScope
	}
	
	// No assignment, just process normally
	p.processPipe(pipe, file, line, scope, templateRefs, fieldRefs)
	return scope
}

// resolveRangeType resolves the element type for a range expression
func (p *ContextAwareParser) resolveRangeType(pipe *parse.PipeNode, file string, line int, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) *ResolvedType {
	if pipe == nil || len(pipe.Cmds) == 0 {
		return nil
	}
	
	// Get the collection type
	collectionType := p.resolveCommandType(pipe.Cmds[0], file, line, scope, templateRefs, fieldRefs)
	if collectionType == nil {
		return nil
	}
	
	// Extract element type from slice/array
	if strings.HasPrefix(collectionType.Name, "[]") {
		elementTypeName := strings.TrimPrefix(collectionType.Name, "[]")
		elementTypeName = strings.TrimPrefix(elementTypeName, "*") // Handle []*Type
		return p.resolver.GetType(elementTypeName)
	}
	
	// For now, we don't handle maps - could be extended
	return nil
}

// resolvePipeType resolves the type of a pipeline expression
func (p *ContextAwareParser) resolvePipeType(pipe *parse.PipeNode, file string, line int, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) *ResolvedType {
	if pipe == nil || len(pipe.Cmds) == 0 {
		return nil
	}
	
	return p.resolveCommandType(pipe.Cmds[0], file, line, scope, templateRefs, fieldRefs)
}

// resolveCommandType resolves the return type of a command
func (p *ContextAwareParser) resolveCommandType(cmd *parse.CommandNode, file string, line int, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) *ResolvedType {
	if cmd == nil || len(cmd.Args) == 0 {
		return nil
	}
	
	// Process the first argument to determine type
	switch arg := cmd.Args[0].(type) {
	case *parse.FieldNode:
		return p.resolveFieldType(arg, file, line, scope, templateRefs, fieldRefs)
	case *parse.VariableNode:
		return p.resolveVariableType(arg, scope)
	case *parse.DotNode:
		return scope.DotType
	}
	
	return nil
}

// resolveFieldType resolves the type of a field expression
func (p *ContextAwareParser) resolveFieldType(field *parse.FieldNode, file string, line int, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) *ResolvedType {
	if field == nil || len(field.Ident) == 0 {
		return nil
	}
	
	// Start with the first identifier
	firstIdent := field.Ident[0]
	
	// Check if it's a variable reference
	if strings.HasPrefix(firstIdent, "$") {
		varType := scope.LookupVariable(firstIdent)
		if varType == nil {
			return nil
		}
		return p.resolveChainFromType(varType, field.Ident[1:], file, line, fieldRefs)
	}
	
	// Check if it's a controller reference
	if controller, ok := p.controllers[firstIdent]; ok {
		// Controller method reference
		if len(field.Ident) >= 2 {
			methodName := field.Ident[1]
			ref := TemplateReference{
				File:       file,
				Line:       line,
				Controller: firstIdent,
				Method:     methodName,
				Full:       fmt.Sprintf("%s.%s", firstIdent, methodName),
			}
			*templateRefs = append(*templateRefs, ref)
			
			// Find the method's return type
			methodType := p.findMethodReturnType(controller, methodName)
			if methodType != nil && len(field.Ident) > 2 {
				// Continue resolving the chain
				return p.resolveChainFromType(methodType, field.Ident[2:], file, line, fieldRefs)
			}
			return methodType
		}
		return nil
	}
	
	// It's a field on the current context (.)
	if scope.DotType != nil {
		return p.resolveChainFromType(scope.DotType, field.Ident, file, line, fieldRefs)
	}
	
	return nil
}

// resolveVariableType resolves the type of a variable
func (p *ContextAwareParser) resolveVariableType(variable *parse.VariableNode, scope *TemplateScope) *ResolvedType {
	if variable == nil || len(variable.Ident) == 0 {
		return nil
	}
	
	varName := "$" + strings.Join(variable.Ident, ".")
	return scope.LookupVariable(varName)
}

// resolveChainFromType resolves a chain of field accesses from a known type
func (p *ContextAwareParser) resolveChainFromType(startType *ResolvedType, chain []string, file string, line int, fieldRefs *[]FieldReference) *ResolvedType {
	currentType := startType
	
	for _, fieldName := range chain {
		if currentType == nil {
			break
		}
		
		// Record field reference
		ref := FieldReference{
			File:       file,
			Line:       line,
			Expression: fieldName,
			Fields:     []string{fieldName},
			Context:    currentType.Name,
		}
		*fieldRefs = append(*fieldRefs, ref)
		
		// Look up the field type
		if fieldVar, ok := currentType.Fields[fieldName]; ok {
			// Get the type name from the types.Var
			typeName := p.resolver.GetTypeNameFromType(fieldVar.Type())
			if typeName != "" {
				currentType = p.resolver.GetType(typeName)
			} else {
				return nil
			}
		} else if _, ok := currentType.Methods[fieldName]; ok {
			// It's a method call - we need better tracking of method return types
			// For now, we can't resolve further without proper method signature tracking
			return nil
		} else {
			// Field/method not found
			return nil
		}
	}
	
	return currentType
}

// findMethodReturnType finds the return type of a controller method
func (p *ContextAwareParser) findMethodReturnType(controller *ControllerInfo, methodName string) *ResolvedType {
	// Enhancement: This would need to track actual return types from AST analysis.
	// Currently we don't parse method signatures deeply enough to know return types.
	// This is acceptable as most template validation works without this information.
	// Future work: Extend ControllerInfo to include method signatures from AST.
	return nil
}

// processCommand processes a command node for references
func (p *ContextAwareParser) processCommand(cmd *parse.CommandNode, file string, line int, scope *TemplateScope, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if cmd == nil {
		return
	}
	
	for _, arg := range cmd.Args {
		switch a := arg.(type) {
		case *parse.FieldNode:
			p.resolveFieldType(a, file, line, scope, templateRefs, fieldRefs)
		case *parse.VariableNode:
			// Just resolve type, don't need to track anything
			p.resolveVariableType(a, scope)
		case *parse.DotNode:
			// Current context reference
			if scope.DotType != nil && len(cmd.Args) > 1 {
				// Pipe operations like {{. | method}} are not yet fully supported.
				// This doesn't affect basic template validation as most templates
				// use direct field access rather than complex pipe operations.
				// Future enhancement: Parse and validate pipe operation chains.
			}
		}
	}
}