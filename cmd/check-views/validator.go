package main

import (
	"fmt"
	"strings"
)

// Validate checks all template references against discovered controllers
func Validate(controllers []ControllerInfo, refs []TemplateReference) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Build a map of controller methods for quick lookup
	methodMap := buildMethodMap(controllers)

	// Track statistics
	templateFiles := make(map[string]bool)
	validCount := 0

	// Validate each reference
	for _, ref := range refs {
		templateFiles[ref.File] = true

		// Check if the controller exists
		methods, controllerExists := methodMap[ref.Controller]
		
		if !controllerExists {
			// Controller doesn't exist
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				File:       ref.File,
				Line:       ref.Line,
				Reference:  ref.Full,
				Problem:    fmt.Sprintf("Controller '%s' not found", ref.Controller),
				Suggestion: findSimilarController(ref.Controller, methodMap),
			})
			continue
		}

		// Check if the method exists
		methodExists := false
		for _, method := range methods {
			if method == ref.Method {
				methodExists = true
				validCount++
				break
			}
		}

		if !methodExists {
			// Method doesn't exist
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				File:       ref.File,
				Line:       ref.Line,
				Reference:  ref.Full,
				Problem:    fmt.Sprintf("Method '%s' not found in controller '%s'", ref.Method, ref.Controller),
				Suggestion: findSimilarMethod(ref.Method, methods),
			})
		}
	}

	// Update statistics
	result.Summary = Statistics{
		TotalTemplates:  len(templateFiles),
		TotalReferences: len(refs),
		ValidReferences: validCount,
		Errors:          len(result.Errors),
	}

	return result
}

// buildMethodMap creates a map of controller prefix to methods
func buildMethodMap(controllers []ControllerInfo) map[string][]string {
	methodMap := make(map[string][]string)
	
	for _, controller := range controllers {
		methodMap[controller.Prefix] = controller.Methods
	}
	
	return methodMap
}

// findSimilarController suggests a similar controller name
func findSimilarController(invalid string, methodMap map[string][]string) string {
	var candidates []string
	
	for controller := range methodMap {
		candidates = append(candidates, controller)
	}
	
	similar := findMostSimilar(invalid, candidates)
	if similar != "" {
		return fmt.Sprintf("Did you mean '%s'?", similar)
	}
	
	return "Available controllers: " + strings.Join(candidates, ", ")
}

// findSimilarMethod suggests a similar method name
func findSimilarMethod(invalid string, methods []string) string {
	similar := findMostSimilar(invalid, methods)
	if similar != "" {
		return fmt.Sprintf("Did you mean '%s'?", similar)
	}
	
	// If no similar method found, suggest common patterns
	if strings.HasPrefix(invalid, "Get") {
		// Look for methods that might be getters
		var getters []string
		for _, m := range methods {
			if strings.HasPrefix(m, "Get") || strings.HasPrefix(m, "All") || strings.HasPrefix(m, "List") {
				getters = append(getters, m)
			}
		}
		if len(getters) > 0 {
			return "Available getters: " + strings.Join(getters[:min(3, len(getters))], ", ")
		}
	}
	
	if strings.HasPrefix(invalid, "Is") || strings.HasPrefix(invalid, "Has") {
		// Look for boolean methods
		var booleans []string
		for _, m := range methods {
			if strings.HasPrefix(m, "Is") || strings.HasPrefix(m, "Has") || strings.HasPrefix(m, "Can") {
				booleans = append(booleans, m)
			}
		}
		if len(booleans) > 0 {
			return "Available checks: " + strings.Join(booleans[:min(3, len(booleans))], ", ")
		}
	}
	
	// Show first few available methods
	if len(methods) > 0 {
		shown := methods[:min(5, len(methods))]
		return "Available methods: " + strings.Join(shown, ", ")
	}
	
	return ""
}

// findMostSimilar finds the most similar string using Levenshtein distance
func findMostSimilar(target string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	
	bestMatch := ""
	bestDistance := len(target) + 10 // High initial value
	
	targetLower := strings.ToLower(target)
	
	for _, candidate := range candidates {
		// Calculate distance
		distance := levenshteinDistance(targetLower, strings.ToLower(candidate))
		
		// Consider it a match if distance is small enough
		if distance < bestDistance && distance <= len(target)/2+1 {
			bestDistance = distance
			bestMatch = candidate
		}
		
		// Also check if one contains the other
		if strings.Contains(strings.ToLower(candidate), targetLower) ||
		   strings.Contains(targetLower, strings.ToLower(candidate)) {
			if len(candidate) < len(bestMatch) || bestMatch == "" {
				bestMatch = candidate
				bestDistance = 1
			}
		}
	}
	
	return bestMatch
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	// Create a 2D slice for dynamic programming
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}
	
	// Initialize first column and row
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}
	
	// Fill in the matrix
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
	
	minVal := nums[0]
	for _, n := range nums[1:] {
		if n < minVal {
			minVal = n
		}
	}
	return minVal
}