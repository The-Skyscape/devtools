package main

import (
	"fmt"
	"log"
	"strings"
)

// EnhancedValidator uses TypeResolver for accurate validation
type EnhancedValidator struct {
	resolver *TypeResolver
	context  *TemplateContextTracker
	verbose  bool
}

// NewEnhancedValidator creates a new validator with type resolution
func NewEnhancedValidator(dir string, verbose bool) (*EnhancedValidator, error) {
	// Create type resolver
	resolver, err := NewTypeResolver(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create type resolver: %w", err)
	}

	// Create template context tracker
	context := NewTemplateContextTracker(resolver)

	// Analyze controllers to understand what they pass to templates
	if err := context.AnalyzeControllers(dir); err != nil {
		if verbose {
			log.Printf("Warning: Failed to analyze controllers: %v", err)
		}
	}

	// Analyze template includes
	if err := context.AnalyzeTemplateIncludes(dir); err != nil {
		if verbose {
			log.Printf("Warning: Failed to analyze template includes: %v", err)
		}
	}

	return &EnhancedValidator{
		resolver: resolver,
		context:  context,
		verbose:  verbose,
	}, nil
}

// ValidateFieldReference validates a field reference using full type resolution
func (v *EnhancedValidator) ValidateFieldReference(ref FieldReference) *FieldValidationError {
	if len(ref.Fields) == 0 {
		return nil
	}

	// Get template context to understand what type is available
	templateContext := v.context.GetTemplateContext(ref.File)

	// If we don't know the context, we cannot validate fields
	// This is expected - we only validate what we can trace through AST
	if templateContext == nil {
		if v.verbose {
			log.Printf("No context for template %s, skipping field validation", ref.File)
		}
		// Return nil - no error, just can't validate without context
		return nil
	}

	// Validate against the specific type
	return v.validateAgainstType(ref, templateContext)
}

// validateAgainstType validates a field reference against a specific type
func (v *EnhancedValidator) validateAgainstType(ref FieldReference, contextType *ResolvedType) *FieldValidationError {
	if contextType == nil || len(ref.Fields) == 0 {
		return nil
	}

	currentType := contextType
	fieldPath := ref.Fields

	// Walk through the field path
	for i, fieldName := range fieldPath {
		// Check if this field exists on the current type
		if !currentType.HasField(fieldName) {
			// Field not found - generate error
			problem := fmt.Sprintf("Field '%s' not found on type %s", fieldName, currentType.Name)
			suggestion := v.findSimilarField(fieldName, currentType)

			return &FieldValidationError{
				File:       ref.File,
				Line:       ref.Line,
				Expression: ref.Expression,
				Fields:     fieldPath[:i+1],
				Context:    ref.Context,
				Problem:    problem,
				Suggestion: suggestion,
			}
		}

		// If this isn't the last field, get the type of this field
		if i < len(fieldPath)-1 {
			fieldType := currentType.GetFieldType(fieldName)
			if fieldType == nil {
				// Can't resolve further
				if v.verbose {
					log.Printf("Can't resolve type of field %s.%s", currentType.Name, fieldName)
				}
				return nil // Be lenient
			}

			// Try to resolve to our known types
			nextType := v.resolver.resolveType(fieldType)
			if nextType == nil {
				// Can't continue resolution
				if v.verbose {
					log.Printf("Unknown type for field %s.%s", currentType.Name, fieldName)
				}
				return nil // Be lenient
			}

			currentType = nextType
		}
	}

	// All fields validated successfully
	return nil
}

// validateAgainstAllTypesStrict is a stricter version for unknown contexts
func (v *EnhancedValidator) validateAgainstAllTypesStrict(ref FieldReference) *FieldValidationError {
	if len(ref.Fields) == 0 {
		return nil
	}

	firstField := ref.Fields[0]
	
	// For templates with "issue" in the name, validate against Issue type
	if strings.Contains(strings.ToLower(ref.File), "issue") {
		if issueType := v.resolver.GetType("Issue"); issueType != nil {
			if !issueType.HasField(firstField) && issueType.Methods[firstField] == nil {
				problem := fmt.Sprintf("Field '%s' not found on type Issue", firstField)
				suggestion := v.findSimilarField(firstField, issueType)
				return &FieldValidationError{
					File:       ref.File,
					Line:       ref.Line,
					Expression: ref.Expression,
					Fields:     ref.Fields,
					Context:    "Issue (inferred)",
					Problem:    problem,
					Suggestion: suggestion,
				}
			}
		}
	}
	
	// Fall back to the original lenient check
	return v.validateAgainstAllTypes(ref)
}

// validateAgainstAllTypes checks if the field exists on any known type
func (v *EnhancedValidator) validateAgainstAllTypes(ref FieldReference) *FieldValidationError {
	if len(ref.Fields) == 0 {
		return nil
	}

	firstField := ref.Fields[0]

	// Check all types to see if any have this field
	var typesWithField []string
	var typesWithMethod []string

	for typeName, rt := range v.resolver.typeMap {
		// Skip duplicates (we store both qualified and unqualified names)
		if strings.Contains(typeName, ".") {
			continue
		}

		if rt.HasField(firstField) {
			if _, isField := rt.Fields[firstField]; isField {
				typesWithField = append(typesWithField, typeName)
			} else {
				typesWithMethod = append(typesWithMethod, typeName)
			}
		}
	}

	// If found on any type, consider it valid (we don't know the context)
	if len(typesWithField) > 0 || len(typesWithMethod) > 0 {
		return nil
	}

	// Not found on any type - this is an error
	problem := fmt.Sprintf("'%s' not found on any known type", firstField)
	suggestion := v.findSimilarFieldAcrossTypes(firstField)

	return &FieldValidationError{
		File:       ref.File,
		Line:       ref.Line,
		Expression: ref.Expression,
		Fields:     ref.Fields,
		Context:    ref.Context,
		Problem:    problem,
		Suggestion: suggestion,
	}
}

// findSimilarField finds similar field names on a type for suggestions
func (v *EnhancedValidator) findSimilarField(fieldName string, rt *ResolvedType) string {
	if rt == nil {
		return ""
	}

	// Check for common mistakes

	// 1. Get prefix that should be removed
	if strings.HasPrefix(fieldName, "Get") && len(fieldName) > 3 {
		withoutGet := strings.TrimPrefix(fieldName, "Get")
		if rt.HasField(withoutGet) {
			return fmt.Sprintf("Did you mean '%s'?", withoutGet)
		}
	}

	// 2. Case sensitivity issues
	lowerField := strings.ToLower(fieldName)
	for name := range rt.Fields {
		if strings.ToLower(name) == lowerField {
			return fmt.Sprintf("Did you mean '%s'? (case sensitive)", name)
		}
	}
	for name := range rt.Methods {
		if strings.ToLower(name) == lowerField {
			return fmt.Sprintf("Did you mean '%s'? (case sensitive)", name)
		}
	}

	// 3. Similar names (simple edit distance)
	suggestions := []string{}
	for name := range rt.Fields {
		if isSimular(fieldName, name) {
			suggestions = append(suggestions, name)
		}
	}
	for name := range rt.Methods {
		if isSimular(fieldName, name) {
			suggestions = append(suggestions, name)
		}
	}

	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			return fmt.Sprintf("Did you mean '%s'?", suggestions[0])
		}
		return fmt.Sprintf("Did you mean one of: %s?", strings.Join(suggestions, ", "))
	}

	return ""
}

// findSimilarFieldAcrossTypes finds similar fields across all types
func (v *EnhancedValidator) findSimilarFieldAcrossTypes(fieldName string) string {
	suggestions := make(map[string][]string)

	// Check all types
	for typeName, rt := range v.resolver.typeMap {
		// Skip qualified names
		if strings.Contains(typeName, ".") {
			continue
		}

		// Check for exact match without Get prefix
		if strings.HasPrefix(fieldName, "Get") {
			withoutGet := strings.TrimPrefix(fieldName, "Get")
			if rt.HasField(withoutGet) {
				suggestions[withoutGet] = append(suggestions[withoutGet], typeName)
			}
		}

		// Check for case-insensitive matches
		lowerField := strings.ToLower(fieldName)
		for name := range rt.Fields {
			if strings.ToLower(name) == lowerField && name != fieldName {
				suggestions[name] = append(suggestions[name], typeName)
			}
		}
		for name := range rt.Methods {
			if strings.ToLower(name) == lowerField && name != fieldName {
				suggestions[name] = append(suggestions[name], typeName)
			}
		}
	}

	if len(suggestions) > 0 {
		// Format suggestions
		var parts []string
		for field, types := range suggestions {
			if len(types) == 1 {
				parts = append(parts, fmt.Sprintf("'%s' (on %s)", field, types[0]))
			} else {
				parts = append(parts, fmt.Sprintf("'%s' (on %s)", field, strings.Join(types, ", ")))
			}
		}

		if len(parts) == 1 {
			return "Did you mean " + parts[0] + "?"
		}
		return "Did you mean one of: " + strings.Join(parts, ", ") + "?"
	}

	return ""
}

// isSimular checks if two strings are similar (simple heuristic)
func isSimular(a, b string) bool {
	// Very simple similarity check
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	// Check if one is a substring of the other
	if strings.Contains(strings.ToLower(a), strings.ToLower(b)) ||
		strings.Contains(strings.ToLower(b), strings.ToLower(a)) {
		return true
	}

	// Check edit distance for short strings
	if len(a) < 10 && len(b) < 10 {
		dist := simpleEditDistance(strings.ToLower(a), strings.ToLower(b))
		maxLen := len(a)
		if len(b) > maxLen {
			maxLen = len(b)
		}
		// Allow up to 2 edits for short strings
		return dist <= 2 && dist < maxLen/2
	}

	return false
}

// editDistance calculates simple edit distance between two strings
func simpleEditDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Simple Levenshtein distance
	if len(a) > len(b) {
		a, b = b, a
	}

	prev := make([]int, len(a)+1)
	curr := make([]int, len(a)+1)

	for i := range prev {
		prev[i] = i
	}

	for j := 1; j <= len(b); j++ {
		curr[0] = j
		for i := 1; i <= len(a); i++ {
			if a[i-1] == b[j-1] {
				curr[i] = prev[i-1]
			} else {
				curr[i] = min3(prev[i]+1, curr[i-1]+1, prev[i-1]+1)
			}
		}
		prev, curr = curr, prev
	}

	return prev[len(a)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Validate checks controller references (basic validation)
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

// ValidateFieldWithTypes validates a field/method reference by checking all types (fallback)
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
	var typesWithCorrectName []string

	// Check all types to see which ones have this field or method
	for typeName, typeInfo := range types {
		// Check fields
		if _, hasField := typeInfo.Fields[fieldOrMethod]; hasField {
			typesWithField = append(typesWithField, typeName)
		}

		// Check if the correct name exists as a field (without Get prefix)
		if possibleCorrectName != "" {
			if _, hasField := typeInfo.Fields[possibleCorrectName]; hasField {
				typesWithCorrectName = append(typesWithCorrectName, typeName)
			}
		}

		// Check methods
		for _, method := range typeInfo.Methods {
			if method.Name == fieldOrMethod {
				typesWithMethod = append(typesWithMethod, typeName)
			}
			// Check if the correct name exists as a method (without Get prefix)
			if possibleCorrectName != "" && method.Name == possibleCorrectName {
				typesWithCorrectName = append(typesWithCorrectName, typeName)
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
	if len(typesWithCorrectName) > 0 {
		suggestion = fmt.Sprintf("Did you mean '%s'? Found on: %s",
			possibleCorrectName,
			strings.Join(typesWithCorrectName, ", "))
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

// ValidateWithTypes performs validation using all discovered types (fallback)
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

// ValidateWithResolver performs full validation using type resolution
func ValidateWithResolver(dir string, controllers []ControllerInfo, templateRefs []TemplateReference, fieldRefs []FieldReference) (*EnhancedValidationResult, error) {
	// Create enhanced validator
	validator, err := NewEnhancedValidator(dir, verbose)
	if err != nil {
		// Fall back to simple validation
		if verbose {
			log.Printf("Failed to create enhanced validator, using simple validation: %v", err)
		}

		// Discover types the old way
		allTypes, _ := DiscoverAllTypes(dir)
		return ValidateWithTypes(controllers, allTypes, templateRefs, fieldRefs), nil
	}

	result := &EnhancedValidationResult{
		Valid:            true,
		ControllerErrors: []ValidationError{},
		FieldErrors:      []FieldValidationError{},
	}

	// Validate controller references (existing logic)
	controllerResult := Validate(controllers, templateRefs)
	result.ControllerErrors = controllerResult.Errors
	if len(controllerResult.Errors) > 0 {
		result.Valid = false
	}

	// Validate field references with enhanced validator
	validFieldCount := 0
	templateFiles := make(map[string]bool)

	for _, ref := range fieldRefs {
		templateFiles[ref.File] = true

		if err := validator.ValidateFieldReference(ref); err != nil {
			result.Valid = false
			result.FieldErrors = append(result.FieldErrors, *err)
		} else {
			validFieldCount++
		}
	}

	// Track template files
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

	return result, nil
}
