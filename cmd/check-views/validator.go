package main

import (
	"fmt"
	"strings"
)

// Validate checks controller references
func Validate(controllers []ControllerInfo, references []TemplateReference) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}
	
	// Track valid references
	validCount := 0
	
	for _, ref := range references {
		// Check if controller exists
		controllerFound := false
		var foundController *ControllerInfo
		
		for _, ctrl := range controllers {
			if ctrl.Prefix == ref.Controller {
				controllerFound = true
				foundController = &ctrl
				break
			}
		}
		
		if !controllerFound {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				File:      ref.File,
				Line:      ref.Line,
				Reference: ref.Full,
				Problem:   fmt.Sprintf("Controller '%s' not found", ref.Controller),
			})
			continue
		}
		
		// Check if method exists
		methodFound := false
		for _, method := range foundController.Methods {
			if method == ref.Method {
				methodFound = true
				validCount++
				break
			}
		}
		
		if !methodFound {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				File:      ref.File,
				Line:      ref.Line,
				Reference: ref.Full,
				Problem:   fmt.Sprintf("Method '%s' not found in controller '%s'", ref.Method, ref.Controller),
			})
		}
	}
	
	result.Summary = Statistics{
		TotalReferences: len(references),
		ValidReferences: validCount,
		Errors:          len(result.Errors),
	}
	
	return result
}

// ValidateFieldWithTypes validates a field/method reference by checking all types
func ValidateFieldWithTypes(ref FieldReference, types map[string]*TypeInfo) *FieldValidationError {
	if len(ref.Fields) == 0 {
		return nil
	}
	
	// Get the first field/method being accessed
	fieldOrMethod := ref.Fields[0]
	
	// If it starts with Get, it's likely a method that should exist without Get prefix
	possibleCorrectName := ""
	if strings.HasPrefix(fieldOrMethod, "Get") && len(fieldOrMethod) > 3 {
		// GetInitials -> Initials
		possibleCorrectName = strings.TrimPrefix(fieldOrMethod, "Get")
	}
	
	// Track which types have this field/method
	var typesWithField []string
	var typesWithMethod []string
	var typesWithCorrectMethod []string
	
	// Check all types to see which ones have this field or method
	for typeName, typeInfo := range types {
		// Check fields
		if _, hasField := typeInfo.Fields[fieldOrMethod]; hasField {
			typesWithField = append(typesWithField, typeName)
		}
		
		// Check methods
		for _, method := range typeInfo.Methods {
			if method.Name == fieldOrMethod {
				typesWithMethod = append(typesWithMethod, typeName)
			}
			// Check if the correct name exists (without Get prefix)
			if possibleCorrectName != "" && method.Name == possibleCorrectName {
				typesWithCorrectMethod = append(typesWithCorrectMethod, typeName)
			}
		}
	}
	
	// If we found it in any type, it's potentially valid (we just don't know the context)
	if len(typesWithField) > 0 || len(typesWithMethod) > 0 {
		// Valid - exists in at least one type
		return nil
	}
	
	// Not found in any type - this is definitely an error
	problem := fmt.Sprintf("'%s' is not a field or method on any known type", fieldOrMethod)
	suggestion := ""
	
	// If we found the correct name (without Get), suggest it
	if len(typesWithCorrectMethod) > 0 {
		suggestion = fmt.Sprintf("Did you mean '%s'? Found on: %s", 
			possibleCorrectName, 
			strings.Join(typesWithCorrectMethod, ", "))
	}
	
	return &FieldValidationError{
		File:       ref.File,
		Line:       ref.Line,
		Expression: ref.Expression,
		Context:    ref.Context,
		Problem:    problem,
		Suggestion: suggestion,
	}
}

// ValidateFieldSimple validates a field/method reference by checking all models
func ValidateFieldSimple(ref FieldReference, models map[string]*ModelInfo) *FieldValidationError {
	if len(ref.Fields) == 0 {
		return nil
	}
	
	// Get the first field/method being accessed
	fieldOrMethod := ref.Fields[0]
	
	// If it starts with Get, it's likely a method that should exist without Get prefix
	possibleCorrectName := ""
	if strings.HasPrefix(fieldOrMethod, "Get") && len(fieldOrMethod) > 3 {
		// GetInitials -> Initials
		possibleCorrectName = strings.TrimPrefix(fieldOrMethod, "Get")
	}
	
	// Track which models have this field/method
	var modelsWithField []string
	var modelsWithMethod []string
	var modelsWithCorrectMethod []string
	
	// Check all models to see which ones have this field or method
	for modelName, model := range models {
		// Check fields
		if _, hasField := model.Fields[fieldOrMethod]; hasField {
			modelsWithField = append(modelsWithField, modelName)
		}
		
		// Check methods
		for _, method := range model.Methods {
			if method.Name == fieldOrMethod {
				modelsWithMethod = append(modelsWithMethod, modelName)
			}
			// Check if the correct name exists (without Get prefix)
			if possibleCorrectName != "" && method.Name == possibleCorrectName {
				modelsWithCorrectMethod = append(modelsWithCorrectMethod, modelName)
			}
		}
	}
	
	// If we found it in any model, it's potentially valid (we just don't know the context)
	if len(modelsWithField) > 0 || len(modelsWithMethod) > 0 {
		// Valid - exists in at least one model
		return nil
	}
	
	// Not found in any model - this is definitely an error
	problem := fmt.Sprintf("'%s' is not a field or method on any known model", fieldOrMethod)
	suggestion := ""
	
	// If we found the correct name (without Get), suggest it
	if len(modelsWithCorrectMethod) > 0 {
		suggestion = fmt.Sprintf("Did you mean '%s'? Found on: %s", 
			possibleCorrectName, 
			strings.Join(modelsWithCorrectMethod, ", "))
	}
	
	return &FieldValidationError{
		File:       ref.File,
		Line:       ref.Line,
		Expression: ref.Expression,
		Context:    ref.Context,
		Problem:    problem,
		Suggestion: suggestion,
	}
}

// ValidateWithTypes performs validation using all discovered types
func ValidateWithTypes(controllers []ControllerInfo, types map[string]*TypeInfo, templateRefs []TemplateReference, fieldRefs []FieldReference) *EnhancedValidationResult {
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
	
	// Validate field references with enhanced type checking
	validFieldCount := 0
	templateFiles := make(map[string]bool)
	
	for _, ref := range fieldRefs {
		templateFiles[ref.File] = true
		
		// Use enhanced validation with all types
		if err := ValidateFieldWithTypes(ref, types); err != nil {
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
	
	return &result
}

// ValidateWithModelsSimple performs validation without requiring perfect context
func ValidateWithModels(controllers []ControllerInfo, models map[string]*ModelInfo, templateRefs []TemplateReference, fieldRefs []FieldReference) *EnhancedValidationResult {
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
	
	// Validate field references with simple validation
	validFieldCount := 0
	templateFiles := make(map[string]bool)
	
	for _, ref := range fieldRefs {
		templateFiles[ref.File] = true
		
		// Use simple validation that doesn't require context
		if err := ValidateFieldSimple(ref, models); err != nil {
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
	
	return &result
}