package main

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	// Discover controllers
	controllers, err := DiscoverControllers("testdata")
	if err != nil {
		t.Fatalf("Failed to discover controllers: %v", err)
	}

	// Parse templates
	refs, err := ParseTemplates("testdata")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	// Validate
	result := Validate(controllers, refs)

	// Should have errors
	if result.Valid {
		t.Error("Validation should fail due to invalid references")
	}

	// Check for specific errors
	foundNonExistentError := false
	foundInvalidMethodError := false

	for _, err := range result.Errors {
		if err.Reference == "home.NonExistentMethod" {
			foundNonExistentError = true
		}
		if err.Reference == "auth.InvalidMethod" {
			foundInvalidMethodError = true
		}
	}

	if !foundNonExistentError {
		t.Error("Should have error for home.NonExistentMethod")
	}
	// Private methods are not parsed by the regex (lowercase methods not accessible)
	// So we don't expect an error for home.privateMethod
	if !foundInvalidMethodError {
		t.Error("Should have error for auth.InvalidMethod")
	}

	// Valid references should pass
	if result.Summary.ValidReferences == 0 {
		t.Error("Should have some valid references")
	}
}

func TestValidateWithResolver(t *testing.T) {
	// Test with enhanced resolution
	controllers, err := DiscoverControllersEnhanced("testdata")
	if err != nil {
		t.Fatalf("Failed to discover controllers: %v", err)
	}

	refs, err := ParseTemplates("testdata")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	// For field refs, we'll create a simple test
	fieldRefs := []FieldReference{
		{
			File:       "test.html",
			Line:       1,
			Expression: "CurrentUser.Email",
			Fields:     []string{"CurrentUser", "Email"},
		},
	}

	result, err := ValidateWithResolver("testdata", controllers, refs, fieldRefs)
	if err != nil {
		t.Fatalf("Failed to validate with resolver: %v", err)
	}

	// Should have controller errors but not for embedded methods
	hasErrorForCurrentUser := false
	hasErrorForIsAuthenticated := false

	for _, err := range result.ControllerErrors {
		if err.Reference == "auth.CurrentUser" {
			hasErrorForCurrentUser = true
		}
		if err.Reference == "auth.IsAuthenticated" {
			hasErrorForIsAuthenticated = true
		}
	}

	if hasErrorForCurrentUser {
		t.Error("Should NOT have error for auth.CurrentUser (embedded method)")
	}
	if hasErrorForIsAuthenticated {
		t.Error("Should NOT have error for auth.IsAuthenticated (embedded method)")
	}
}

func TestValidateFieldWithTypes(t *testing.T) {
	// Create test types
	types := map[string]*TypeInfo{
		"User": {
			Name: "User",
			Fields: map[string]FieldInfo{
				"Email": {Name: "Email", Type: "string", IsExported: true},
				"Name":  {Name: "Name", Type: "string", IsExported: true},
			},
			Methods: []MethodInfo{
				{Name: "GetInitials", ReturnType: "string"},
			},
		},
	}

	// Test valid field
	ref := FieldReference{
		File:       "test.html",
		Line:       1,
		Expression: ".Email",
		Fields:     []string{"Email"},
	}

	err := ValidateFieldWithTypes(ref, types)
	if err != nil {
		t.Errorf("Should not have error for valid field: %v", err)
	}

	// Test invalid field
	ref = FieldReference{
		File:       "test.html",
		Line:       2,
		Expression: ".InvalidField",
		Fields:     []string{"InvalidField"},
	}

	err = ValidateFieldWithTypes(ref, types)
	if err == nil {
		t.Error("Should have error for invalid field")
	}

	// Test method with Get prefix
	ref = FieldReference{
		File:       "test.html",
		Line:       3,
		Expression: ".GetInitials",
		Fields:     []string{"GetInitials"},
	}

	err = ValidateFieldWithTypes(ref, types)
	if err != nil {
		t.Errorf("Should not have error for valid method: %v", err)
	}

	// Test suggestion for Get prefix mistake
	ref = FieldReference{
		File:       "test.html",
		Line:       4,
		Expression: ".GetEmail",
		Fields:     []string{"GetEmail"},
	}

	err = ValidateFieldWithTypes(ref, types)
	if err == nil {
		t.Error("Should have error for GetEmail")
	}
	// The suggestion should mention Email since it exists
	if err != nil && !strings.Contains(err.Suggestion, "Email") {
		t.Errorf("Should have suggestion mentioning 'Email', got: %s", err.Suggestion)
	}
}

func TestEditDistanceAndSimilarity(t *testing.T) {
	tests := []struct {
		a, b     string
		expected bool
		name     string
	}{
		{"Email", "email", true, "case difference"},
		{"GetUser", "GetUsers", true, "plural"},
		{"Name", "Names", true, "singular/plural"},
		{"User", "Users", true, "close match"},
		{"ABC", "XYZ", false, "completely different"},
		{"Method", "method", true, "case only"},
		{"GetInitials", "Initials", true, "substring"},
	}

	for _, test := range tests {
		result := isSimular(test.a, test.b)
		if result != test.expected {
			t.Errorf("isSimular(%s, %s) = %v, expected %v (%s)",
				test.a, test.b, result, test.expected, test.name)
		}
	}

	// Test edit distance directly
	dist := simpleEditDistance("kitten", "sitting")
	if dist != 3 {
		t.Errorf("Edit distance between 'kitten' and 'sitting' should be 3, got %d", dist)
	}
}
