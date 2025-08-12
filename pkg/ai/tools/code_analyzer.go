package tools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// CodeAnalyzer is a tool for analyzing code
type CodeAnalyzer struct {
	name        string
	description string
}

// NewCodeAnalyzer creates a new code analyzer tool
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		name:        "code_analyzer",
		description: "Analyzes code files for structure, complexity, and potential issues",
	}
}

// Name returns the tool name
func (c *CodeAnalyzer) Name() string {
	return c.name
}

// Description returns the tool description
func (c *CodeAnalyzer) Description() string {
	return c.description
}

// Parameters returns the tool's parameter schema
func (c *CodeAnalyzer) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to analyze",
			},
			"analysis_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of analysis to perform",
				"enum":        []string{"structure", "complexity", "dependencies", "all"},
			},
		},
		"required": []string{"file_path"},
	}
}

// Execute runs the code analyzer
func (c *CodeAnalyzer) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required")
	}
	
	analysisType := "all"
	if at, ok := params["analysis_type"].(string); ok {
		analysisType = at
	}
	
	// Determine file type
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch ext {
	case ".go":
		return c.analyzeGoFile(filePath, analysisType)
	case ".js", ".jsx", ".ts", ".tsx":
		return c.analyzeJavaScriptFile(filePath, analysisType)
	case ".py":
		return c.analyzePythonFile(filePath, analysisType)
	default:
		return c.analyzeGenericFile(filePath, analysisType)
	}
}

// analyzeGoFile analyzes a Go source file
func (c *CodeAnalyzer) analyzeGoFile(filePath string, analysisType string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}
	
	result := make(map[string]interface{})
	
	// Analyze structure
	if analysisType == "structure" || analysisType == "all" {
		result["structure"] = c.analyzeGoStructure(node)
	}
	
	// Analyze complexity
	if analysisType == "complexity" || analysisType == "all" {
		result["complexity"] = c.analyzeGoComplexity(node)
	}
	
	// Analyze dependencies
	if analysisType == "dependencies" || analysisType == "all" {
		result["dependencies"] = c.analyzeGoDependencies(node)
	}
	
	return result, nil
}

// analyzeGoStructure analyzes the structure of a Go file
func (c *CodeAnalyzer) analyzeGoStructure(node *ast.File) map[string]interface{} {
	structure := map[string]interface{}{
		"package":   node.Name.Name,
		"imports":   []string{},
		"functions": []string{},
		"types":     []string{},
		"variables": []string{},
		"constants": []string{},
	}
	
	// Extract imports
	imports := []string{}
	for _, imp := range node.Imports {
		path := imp.Path.Value
		imports = append(imports, strings.Trim(path, `"`))
	}
	structure["imports"] = imports
	
	// Extract declarations
	functions := []string{}
	types := []string{}
	variables := []string{}
	constants := []string{}
	
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			functions = append(functions, d.Name.Name)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					types = append(types, s.Name.Name)
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if d.Tok == token.VAR {
							variables = append(variables, name.Name)
						} else if d.Tok == token.CONST {
							constants = append(constants, name.Name)
						}
					}
				}
			}
		}
	}
	
	structure["functions"] = functions
	structure["types"] = types
	structure["variables"] = variables
	structure["constants"] = constants
	
	return structure
}

// analyzeGoComplexity calculates cyclomatic complexity for Go code
func (c *CodeAnalyzer) analyzeGoComplexity(node *ast.File) map[string]interface{} {
	complexity := map[string]interface{}{
		"cyclomatic_complexity": 0,
		"line_count":            0,
		"function_count":        0,
		"max_function_length":   0,
	}
	
	// Count functions and calculate complexity
	functionCount := 0
	totalComplexity := 0
	maxFunctionLength := 0
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			functionCount++
			length := c.getFunctionLength(x)
			if length > maxFunctionLength {
				maxFunctionLength = length
			}
			totalComplexity += c.calculateCyclomaticComplexity(x)
		}
		return true
	})
	
	complexity["cyclomatic_complexity"] = totalComplexity
	complexity["function_count"] = functionCount
	complexity["max_function_length"] = maxFunctionLength
	
	return complexity
}

// calculateCyclomaticComplexity calculates the cyclomatic complexity of a function
func (c *CodeAnalyzer) calculateCyclomaticComplexity(fn *ast.FuncDecl) int {
	complexity := 1 // Base complexity
	
	ast.Inspect(fn, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		}
		return true
	})
	
	return complexity
}

// getFunctionLength calculates the number of lines in a function
func (c *CodeAnalyzer) getFunctionLength(fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 0
	}
	// This is a simplified calculation
	return len(fn.Body.List)
}

// analyzeGoDependencies analyzes the dependencies of a Go file
func (c *CodeAnalyzer) analyzeGoDependencies(node *ast.File) map[string]interface{} {
	deps := map[string]interface{}{
		"imports":          []string{},
		"standard_library": []string{},
		"third_party":      []string{},
		"local":            []string{},
	}
	
	stdLib := []string{}
	thirdParty := []string{}
	local := []string{}
	
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		
		// Categorize imports
		if !strings.Contains(path, ".") {
			stdLib = append(stdLib, path)
		} else if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
			local = append(local, path)
		} else {
			thirdParty = append(thirdParty, path)
		}
	}
	
	deps["standard_library"] = stdLib
	deps["third_party"] = thirdParty
	deps["local"] = local
	
	return deps
}

// analyzeJavaScriptFile analyzes a JavaScript/TypeScript file
func (c *CodeAnalyzer) analyzeJavaScriptFile(filePath string, analysisType string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	result := map[string]interface{}{
		"language":   "javascript",
		"file_path":  filePath,
		"line_count": strings.Count(string(content), "\n") + 1,
		"size_bytes": len(content),
	}
	
	// Basic analysis for JavaScript files
	// In a real implementation, would use a proper JavaScript parser
	
	return result, nil
}

// analyzePythonFile analyzes a Python file
func (c *CodeAnalyzer) analyzePythonFile(filePath string, analysisType string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	result := map[string]interface{}{
		"language":   "python",
		"file_path":  filePath,
		"line_count": strings.Count(string(content), "\n") + 1,
		"size_bytes": len(content),
	}
	
	// Basic analysis for Python files
	// In a real implementation, would use a proper Python parser
	
	return result, nil
}

// analyzeGenericFile performs generic analysis on any file
func (c *CodeAnalyzer) analyzeGenericFile(filePath string, analysisType string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	result := map[string]interface{}{
		"file_path":  filePath,
		"line_count": strings.Count(string(content), "\n") + 1,
		"size_bytes": len(content),
		"extension":  filepath.Ext(filePath),
	}
	
	return result, nil
}