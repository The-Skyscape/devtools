package main

import (
	"fmt"
	"strings"
)

// EnhancedValidationResult includes both controller and model validation
type EnhancedValidationResult struct {
	Valid            bool
	ControllerErrors []ValidationError
	FieldErrors      []FieldValidationError
	URLErrors        []URLValidationError
	Summary          EnhancedStatistics
}

// FieldValidationError represents a model field validation error
type FieldValidationError struct {
	File       string
	Line       int
	Expression string   // Full expression like ".User.Name"
	Context    string   // The type context where this appears
	Problem    string   // Description of the problem
	Suggestion string   // Suggested fix
}

// EnhancedStatistics includes both controller and field statistics
type EnhancedStatistics struct {
	TotalTemplates       int
	TotalControllerRefs  int
	ValidControllerRefs  int
	TotalFieldRefs       int
	ValidFieldRefs       int
	TotalURLRefs         int
	ValidURLRefs         int
	ControllerErrors     int
	FieldErrors          int
	URLErrors            int
}

// ValidateWithModels performs comprehensive validation including model fields
func ValidateWithModels(controllers []ControllerInfo, models map[string]*ModelInfo, templateRefs []TemplateReference, fieldRefs []FieldReference) EnhancedValidationResult {
	result := EnhancedValidationResult{
		Valid:            true,
		ControllerErrors: []ValidationError{},
		FieldErrors:      []FieldValidationError{},
	}
	
	// Validate controller references (reuse existing logic)
	controllerResult := Validate(controllers, templateRefs)
	result.ControllerErrors = controllerResult.Errors
	if len(controllerResult.Errors) > 0 {
		result.Valid = false
	}
	
	// Validate field references
	validFieldCount := 0
	templateFiles := make(map[string]bool)
	
	for _, ref := range fieldRefs {
		templateFiles[ref.File] = true
		
		// Skip if context is unknown
		if ref.Context == "" || ref.Context == "root" || ref.Context == "controller_result" {
			// Can't validate without knowing the type
			continue
		}
		
		// Validate the field chain
		if err := validateFieldChain(ref, models); err != nil {
			result.Valid = false
			result.FieldErrors = append(result.FieldErrors, *err)
		} else {
			validFieldCount++
		}
	}
	
	// Also track template files from controller refs
	for _, ref := range templateRefs {
		templateFiles[ref.File] = true
	}
	
	// Update statistics
	result.Summary = EnhancedStatistics{
		TotalTemplates:      len(templateFiles),
		TotalControllerRefs: len(templateRefs),
		ValidControllerRefs: controllerResult.Summary.ValidReferences,
		TotalFieldRefs:      len(fieldRefs),
		ValidFieldRefs:      validFieldCount,
		ControllerErrors:    len(result.ControllerErrors),
		FieldErrors:         len(result.FieldErrors),
	}
	
	return result
}

// validateFieldChain validates a chain of field accesses
func validateFieldChain(ref FieldReference, models map[string]*ModelInfo) *FieldValidationError {
	if len(ref.Fields) == 0 {
		return nil
	}
	
	// Start with the context type
	currentType := ref.Context
	
	// Handle slice context (from range)
	if strings.HasPrefix(currentType, "[]") {
		currentType = strings.TrimPrefix(currentType, "[]")
	}
	
	// Walk through the field chain
	for i, fieldName := range ref.Fields {
		// Find the model for the current type
		model, exists := models[currentType]
		if !exists {
			// Type not found - could be a built-in type or unrecognized
			if isBuiltInType(currentType) {
				// Can't validate fields on built-in types
				return nil
			}
			
			return &FieldValidationError{
				File:       ref.File,
				Line:       ref.Line,
				Expression: ref.Expression,
				Context:    ref.Context,
				Problem:    fmt.Sprintf("Unknown type '%s'", currentType),
				Suggestion: findSimilarModelType(currentType, models),
			}
		}
		
		// Check if the field exists
		fieldInfo, fieldExists := model.Fields[fieldName]
		if !fieldExists {
			// Check if it's a method
			methodFound := false
			var methodReturnType string
			
			for _, method := range model.Methods {
				if method.Name == fieldName {
					methodFound = true
					methodReturnType = method.ReturnType
					break
				}
			}
			
			if methodFound {
				// It's a method - update current type for next iteration
				currentType = cleanTypeName(methodReturnType)
				continue
			}
			
			// Field/method not found
			return &FieldValidationError{
				File:       ref.File,
				Line:       ref.Line,
				Expression: ref.Expression,
				Context:    ref.Context,
				Problem:    fmt.Sprintf("Field or method '%s' not found in type '%s'", fieldName, currentType),
				Suggestion: findSimilarField(fieldName, model),
			}
		}
		
		// Field found - check if we can continue the chain
		if i < len(ref.Fields)-1 {
			// More fields to process - update current type
			currentType = cleanTypeName(fieldInfo.Type)
			
			// Handle pointer types
			if fieldInfo.IsPtr {
				currentType = strings.TrimPrefix(currentType, "*")
			}
			
			// Handle slice types
			if fieldInfo.IsSlice {
				currentType = strings.TrimPrefix(currentType, "[]")
			}
		}
	}
	
	return nil
}

// cleanTypeName removes package qualifiers and pointer indicators
func cleanTypeName(typeName string) string {
	// Remove pointer prefix
	typeName = strings.TrimPrefix(typeName, "*")
	
	// Remove slice prefix
	typeName = strings.TrimPrefix(typeName, "[]")
	
	// Remove package qualifier
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		typeName = parts[len(parts)-1]
	}
	
	return typeName
}

// isBuiltInType checks if a type is a Go built-in type
func isBuiltInType(typeName string) bool {
	builtIns := map[string]bool{
		"string":     true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"float32":    true,
		"float64":    true,
		"bool":       true,
		"byte":       true,
		"rune":       true,
		"error":      true,
		"interface":  true,
		"Time":       true,  // time.Time
		"Duration":   true,  // time.Duration
	}
	
	return builtIns[typeName]
}

// findSimilarModelType suggests a similar model type name
func findSimilarModelType(invalid string, models map[string]*ModelInfo) string {
	var candidates []string
	for typeName := range models {
		candidates = append(candidates, typeName)
	}
	
	similar := findMostSimilar(invalid, candidates)
	if similar != "" {
		return fmt.Sprintf("Did you mean '%s'?", similar)
	}
	
	return "Available types: " + strings.Join(candidates[:min(5, len(candidates))], ", ")
}

// findSimilarField suggests a similar field or method name
func findSimilarField(invalid string, model *ModelInfo) string {
	var candidates []string
	
	// Add fields
	for fieldName, field := range model.Fields {
		if !strings.HasPrefix(fieldName, "*") && field.IsExported {
			candidates = append(candidates, fieldName)
		}
	}
	
	// Add methods
	for _, method := range model.Methods {
		candidates = append(candidates, method.Name)
	}
	
	similar := findMostSimilar(invalid, candidates)
	if similar != "" {
		return fmt.Sprintf("Did you mean '%s'?", similar)
	}
	
	// Show available fields and methods
	var fields []string
	var methods []string
	
	for fieldName, field := range model.Fields {
		if !strings.HasPrefix(fieldName, "*") && field.IsExported {
			fields = append(fields, fieldName)
		}
	}
	
	for _, method := range model.Methods {
		methods = append(methods, method.Name)
	}
	
	suggestion := ""
	if len(fields) > 0 {
		shown := fields[:min(3, len(fields))]
		suggestion += "Available fields: " + strings.Join(shown, ", ")
	}
	
	if len(methods) > 0 {
		shown := methods[:min(3, len(methods))]
		if suggestion != "" {
			suggestion += "; "
		}
		suggestion += "Available methods: " + strings.Join(shown, ", ")
	}
	
	return suggestion
}