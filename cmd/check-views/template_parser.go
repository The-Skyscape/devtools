package main

import (
	"fmt"
	"strings"
	"text/template/parse"
)

// parseTemplateTree parses a template tree and extracts references
func parseTemplateTree(tree *parse.Tree, relPath string, controllers []ControllerInfo, types map[string]*TypeInfo) ([]TemplateReference, []FieldReference) {
	var templateRefs []TemplateReference
	var fieldRefs []FieldReference
	
	if tree != nil && tree.Root != nil {
		walkNode(tree.Root, relPath, controllers, types, &templateRefs, &fieldRefs)
	}
	
	return templateRefs, fieldRefs
}

// walkNode recursively walks the template AST and extracts references
func walkNode(node parse.Node, file string, controllers []ControllerInfo, types map[string]*TypeInfo, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if node == nil {
		return
	}
	
	switch n := node.(type) {
	case *parse.ActionNode:
		if n.Pipe != nil {
			processPipe(n.Pipe, file, n.Line, controllers, types, templateRefs, fieldRefs)
		}
		
	case *parse.IfNode:
		if n.Pipe != nil {
			processPipe(n.Pipe, file, n.Line, controllers, types, templateRefs, fieldRefs)
		}
		if n.List != nil {
			walkNode(n.List, file, controllers, types, templateRefs, fieldRefs)
		}
		if n.ElseList != nil {
			walkNode(n.ElseList, file, controllers, types, templateRefs, fieldRefs)
		}
		
	case *parse.RangeNode:
		if n.Pipe != nil {
			processPipe(n.Pipe, file, n.Line, controllers, types, templateRefs, fieldRefs)
		}
		if n.List != nil {
			walkNode(n.List, file, controllers, types, templateRefs, fieldRefs)
		}
		if n.ElseList != nil {
			walkNode(n.ElseList, file, controllers, types, templateRefs, fieldRefs)
		}
		
	case *parse.WithNode:
		if n.Pipe != nil {
			processPipe(n.Pipe, file, n.Line, controllers, types, templateRefs, fieldRefs)
		}
		if n.List != nil {
			walkNode(n.List, file, controllers, types, templateRefs, fieldRefs)
		}
		if n.ElseList != nil {
			walkNode(n.ElseList, file, controllers, types, templateRefs, fieldRefs)
		}
		
	case *parse.ListNode:
		if n != nil && n.Nodes != nil {
			for _, child := range n.Nodes {
				walkNode(child, file, controllers, types, templateRefs, fieldRefs)
			}
		}
	}
}

// processPipe processes a pipe for references
func processPipe(pipe *parse.PipeNode, file string, line int, controllers []ControllerInfo, types map[string]*TypeInfo, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if pipe == nil {
		return
	}
	
	for _, cmd := range pipe.Cmds {
		processCommand(cmd, file, line, controllers, types, templateRefs, fieldRefs)
	}
}

// processCommand processes a command for references
func processCommand(cmd *parse.CommandNode, file string, line int, controllers []ControllerInfo, types map[string]*TypeInfo, templateRefs *[]TemplateReference, fieldRefs *[]FieldReference) {
	if cmd == nil || len(cmd.Args) == 0 {
		return
	}
	
	for _, arg := range cmd.Args {
		switch a := arg.(type) {
		case *parse.FieldNode:
			if len(a.Ident) > 0 {
				// Check if it's a controller reference
				if len(a.Ident) >= 2 && isControllerName(a.Ident[0], controllers) {
					// Controller method reference
					ref := TemplateReference{
						File:       file,
						Line:       line,
						Controller: a.Ident[0],
						Method:     a.Ident[1],
						Full:       fmt.Sprintf("%s.%s", a.Ident[0], a.Ident[1]),
					}
					*templateRefs = append(*templateRefs, ref)
					
					// If there are more fields after the method, track them as field references
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
					// Regular field/method access - track it
					fieldRef := FieldReference{
						File:       file,
						Line:       line,
						Expression: "." + strings.Join(a.Ident, "."),
						Fields:     a.Ident,
						Context:    "unknown", // We don't track context anymore
					}
					*fieldRefs = append(*fieldRefs, fieldRef)
				}
			}
			
		case *parse.ChainNode:
			// Handle chained field accesses like $var.Field.Method
			if a.Field != nil && len(a.Field) > 0 {
				fieldRef := FieldReference{
					File:       file,
					Line:       line,
					Expression: "chain." + strings.Join(a.Field, "."),
					Fields:     a.Field,
					Context:    "unknown",
				}
				*fieldRefs = append(*fieldRefs, fieldRef)
			}
		}
	}
}