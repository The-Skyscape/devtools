package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RouteInfo represents a discovered HTTP route
type RouteInfo struct {
	Method     string // GET, POST, DELETE, etc.
	Path       string // /repos, /repos/{id}, etc.
	Controller string // Controller that handles this route
	Handler    string // Handler method name
	File       string // File where route is defined
	Line       int    // Line number
}

// URLReference represents a {{host}}/path reference in a template
type URLReference struct {
	File       string // Template file
	Line       int    // Line number
	FullURL    string // Full URL including {{host}}
	Path       string // Just the path part after {{host}}
	Expression string // The full template expression
}

// DiscoverRoutes finds all HTTP routes defined in controllers
func DiscoverRoutes(dir string) ([]RouteInfo, error) {
	var routes []RouteInfo
	
	// Find controllers directory
	controllersDir := filepath.Join(dir, "controllers")
	
	// Walk through all Go files
	err := filepath.WalkDir(controllersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-Go files
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		// Parse the Go file for routes
		fileRoutes, err := parseRoutesFromFile(path)
		if err != nil {
			if verbose {
				fmt.Printf("  ⚠️  Error parsing routes from %s: %v\n", path, err)
			}
			return nil
		}
		
		routes = append(routes, fileRoutes...)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to discover routes: %w", err)
	}
	
	return routes, nil
}

// parseRoutesFromFile extracts route definitions from a Go file
func parseRoutesFromFile(filePath string) ([]RouteInfo, error) {
	var routes []RouteInfo
	
	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	
	// Look for http.Handle calls
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for call expressions
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		
		// Check if it's http.Handle or http.HandleFunc
		if route := extractHTTPHandle(call, fset); route != nil {
			route.File = filePath
			routes = append(routes, *route)
		}
		
		// Check if it's app.Serve or app.ProtectFunc
		if route := extractAppRoute(call, fset); route != nil {
			route.File = filePath
			routes = append(routes, *route)
		}
		
		return true
	})
	
	return routes, nil
}

// extractHTTPHandle extracts route info from http.Handle calls
func extractHTTPHandle(call *ast.CallExpr, fset *token.FileSet) *RouteInfo {
	// Check if it's http.Handle or http.HandleFunc
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	
	// Check package name
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "http" {
		return nil
	}
	
	// Check method name
	if sel.Sel.Name != "Handle" && sel.Sel.Name != "HandleFunc" {
		return nil
	}
	
	// Extract route pattern (first argument)
	if len(call.Args) < 2 {
		return nil
	}
	
	// Get the route pattern
	routePattern := extractStringLiteral(call.Args[0])
	if routePattern == "" {
		return nil
	}
	
	// Parse method and path from pattern (e.g., "GET /repos")
	parts := strings.SplitN(routePattern, " ", 2)
	if len(parts) != 2 {
		// Old style without method, assume GET
		return &RouteInfo{
			Method: "GET",
			Path:   routePattern,
			Line:   fset.Position(call.Pos()).Line,
		}
	}
	
	return &RouteInfo{
		Method: parts[0],
		Path:   parts[1],
		Line:   fset.Position(call.Pos()).Line,
	}
}

// extractAppRoute extracts route info from app.Serve or app.ProtectFunc calls
func extractAppRoute(call *ast.CallExpr, fset *token.FileSet) *RouteInfo {
	// Check if it's a method call
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	
	// Check if the method is Serve or ProtectFunc
	methodName := sel.Sel.Name
	if methodName != "Serve" && methodName != "ProtectFunc" {
		return nil
	}
	
	// Check if the receiver might be 'app'
	if ident, ok := sel.X.(*ast.Ident); ok && ident.Name != "app" {
		return nil
	}
	
	// For app.Serve, first arg is route pattern
	if methodName == "Serve" && len(call.Args) >= 2 {
		routePattern := extractStringLiteral(call.Args[0])
		if routePattern == "" {
			return nil
		}
		
		// Parse method and path
		parts := strings.SplitN(routePattern, " ", 2)
		if len(parts) != 2 {
			return &RouteInfo{
				Method: "GET",
				Path:   routePattern,
				Line:   fset.Position(call.Pos()).Line,
			}
		}
		
		return &RouteInfo{
			Method: parts[0],
			Path:   parts[1],
			Line:   fset.Position(call.Pos()).Line,
		}
	}
	
	// For app.ProtectFunc, first arg is route pattern
	if methodName == "ProtectFunc" && len(call.Args) >= 2 {
		routePattern := extractStringLiteral(call.Args[0])
		if routePattern == "" {
			return nil
		}
		
		// Parse method and path
		parts := strings.SplitN(routePattern, " ", 2)
		if len(parts) != 2 {
			return &RouteInfo{
				Method: "POST", // ProtectFunc usually for POST
				Path:   routePattern,
				Line:   fset.Position(call.Pos()).Line,
			}
		}
		
		return &RouteInfo{
			Method: parts[0],
			Path:   parts[1],
			Line:   fset.Position(call.Pos()).Line,
		}
	}
	
	return nil
}

// extractStringLiteral extracts a string literal from an AST expression
func extractStringLiteral(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			// Remove quotes
			s := e.Value
			if len(s) >= 2 {
				return s[1 : len(s)-1]
			}
		}
	}
	return ""
}

// ParseHostReferences finds all {{host}}/path references in templates
func ParseHostReferences(dir string) ([]URLReference, error) {
	var refs []URLReference
	
	// Find views directory
	viewsDir := filepath.Join(dir, "views")
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		viewsDir = dir
	}
	
	// Regex to match {{host}}/path patterns
	// Matches: {{host}}/path, {{host}}/path/to/resource, href="{{host}}/path", etc.
	hostPattern := regexp.MustCompile(`\{\{host\}\}(/[^"\s<>}]*)`)
	
	// Walk through all HTML files
	err := filepath.WalkDir(viewsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-HTML files
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		
		// Parse the HTML file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		
		// Get relative path for display
		relPath, _ := filepath.Rel(viewsDir, path)
		if relPath == "" {
			relPath = filepath.Base(path)
		}
		
		// Find all {{host}} references
		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			matches := hostPattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 2 {
					if verbose && strings.Contains(match[1], "fake") {
						fmt.Printf("  📍 Found URL reference: %s in %s:%d\n", match[0], relPath, lineNum+1)
					}
					refs = append(refs, URLReference{
						File:       relPath,
						Line:       lineNum + 1,
						FullURL:    match[0],
						Path:       match[1],
						Expression: match[0],
					})
				}
			}
		}
		
		return nil
	})
	
	return refs, err
}

// ValidateURLReferences checks if URL references match actual routes
func ValidateURLReferences(refs []URLReference, routes []RouteInfo) []URLValidationError {
	var errors []URLValidationError
	
	// Build maps for validation
	exactPaths := make(map[string]bool)
	patternPaths := []string{}
	
	for _, route := range routes {
		// Normalize the path
		path := route.Path
		
		// Separate exact paths from patterns
		if strings.Contains(path, "{") {
			// Add as pattern
			pattern := pathToRegex(path)
			patternPaths = append(patternPaths, pattern)
		} else {
			// Add as exact path
			exactPaths[path] = true
		}
	}
	
	// Check each URL reference
	for _, ref := range refs {
		// Normalize the reference path
		refPath := ref.Path
		
		// Remove query parameters and fragments
		if idx := strings.IndexAny(refPath, "?#"); idx != -1 {
			refPath = refPath[:idx]
		}
		
		// Check for exact match
		if exactPaths[refPath] {
			continue
		}
		
		// Check for pattern match
		matched := false
		for _, pattern := range patternPaths {
			if matchesRoutePattern(refPath, pattern) {
				matched = true
				break
			}
		}
		
		// Check for static files
		if !matched && !isStaticFile(refPath) && !isSpecialPath(refPath) {
			if verbose || strings.Contains(refPath, "fake") {
				fmt.Printf("  ⚠️  Invalid URL found: %s in %s:%d (matched=%v, static=%v, special=%v)\n", 
					ref.Path, ref.File, ref.Line, matched, isStaticFile(refPath), isSpecialPath(refPath))
			}
			errors = append(errors, URLValidationError{
				File:       ref.File,
				Line:       ref.Line,
				URL:        ref.FullURL,
				Path:       ref.Path,
				Problem:    fmt.Sprintf("Route '%s' not found in controllers", ref.Path),
				Suggestion: findSimilarRoute(ref.Path, routes),
			})
		}
	}
	
	return errors
}

// URLValidationError represents a URL validation error
type URLValidationError struct {
	File       string
	Line       int
	URL        string
	Path       string
	Problem    string
	Suggestion string
}

// pathToRegex converts a route path with {param} to a regex pattern
func pathToRegex(path string) string {
	// Replace {param} with [^/]+
	pattern := regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(path, `[^/]+`)
	return "^" + pattern + "$"
}

// matchesRoutePattern checks if a URL path matches a route regex pattern
func matchesRoutePattern(urlPath, pattern string) bool {
	matched, err := regexp.MatchString(pattern, urlPath)
	if err != nil {
		return false
	}
	return matched
}

// isStaticFile checks if a path refers to a static file
func isStaticFile(path string) bool {
	// Common static file extensions
	staticExts := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg",
		".ico", ".woff", ".woff2", ".ttf", ".eot", ".pdf",
		".json", ".xml", ".txt", ".webp",
	}
	
	for _, ext := range staticExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	
	// Check for known static directories
	staticDirs := []string{
		"/static/", "/assets/", "/public/", "/dist/", "/build/",
		"/css/", "/js/", "/images/", "/fonts/", "/media/",
	}
	
	for _, dir := range staticDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}
	
	return false
}

// isSpecialPath checks if a path is a special case that shouldn't be validated
func isSpecialPath(path string) bool {
	// Special paths that might be handled differently
	specialPaths := []string{
		"/", // Root path
		"#", // Fragment only
		"javascript:", // JavaScript URLs
	}
	
	for _, special := range specialPaths {
		if path == special || strings.HasPrefix(path, special) {
			return true
		}
	}
	
	// Check for template variables in the path
	if strings.Contains(path, "{{") {
		return true
	}
	
	return false
}

// findSimilarRoute suggests a similar route for a given path
func findSimilarRoute(path string, routes []RouteInfo) string {
	var candidates []string
	
	for _, route := range routes {
		candidates = append(candidates, route.Path)
	}
	
	similar := findMostSimilar(path, candidates)
	if similar != "" {
		return fmt.Sprintf("Did you mean '%s'?", similar)
	}
	
	// Show some available routes
	if len(routes) > 0 {
		shown := []string{}
		limit := 5
		if len(routes) < limit {
			limit = len(routes)
		}
		for i := 0; i < limit; i++ {
			shown = append(shown, routes[i].Path)
		}
		return "Available routes: " + strings.Join(shown, ", ")
	}
	
	return ""
}

// findMostSimilar finds the most similar string from candidates using edit distance
func findMostSimilar(target string, candidates []string) string {
	minDist := 3 // Minimum distance threshold for suggestions
	bestMatch := ""
	
	for _, candidate := range candidates {
		dist := editDistance(target, candidate)
		if dist < minDist {
			minDist = dist
			bestMatch = candidate
		}
	}
	
	return bestMatch
}

// editDistance calculates the Levenshtein distance between two strings
func editDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	// Create distance matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}
	
	// Fill the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of a set of integers
func min(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	m := nums[0]
	for _, n := range nums[1:] {
		if n < m {
			m = n
		}
	}
	return m
}
