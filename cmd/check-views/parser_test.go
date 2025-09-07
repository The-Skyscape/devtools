package main

import (
	"strings"
	"testing"
)

func TestParseTemplates(t *testing.T) {
	refs, err := ParseTemplates("testdata")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	// Should find controller references
	if len(refs) == 0 {
		t.Error("Should find template references")
	}

	// Debug: print all found references
	t.Logf("Found %d references:", len(refs))
	for _, ref := range refs {
		t.Logf("  - %s", ref.Full)
	}

	// Check specific references
	foundWelcome := false
	foundGetUserCount := false
	foundNonExistent := false
	// Private methods (lowercase) are not parsed as they shouldn't be accessible
	foundPrivateMethod := false
	foundIsAuthenticated := false
	foundCurrentUser := false
	foundInvalidMethod := false

	for _, ref := range refs {
		switch ref.Full {
		case "home.Welcome":
			foundWelcome = true
			if ref.Controller != "home" || ref.Method != "Welcome" {
				t.Errorf("Incorrect parsing of home.Welcome: %+v", ref)
			}
		case "home.GetUserCount":
			foundGetUserCount = true
		case "home.NonExistentMethod":
			foundNonExistent = true
		case "home.privateMethod":
			foundPrivateMethod = true
		case "auth.IsAuthenticated":
			foundIsAuthenticated = true
		case "auth.CurrentUser":
			foundCurrentUser = true
		case "auth.InvalidMethod":
			foundInvalidMethod = true
		}
	}

	if !foundWelcome {
		t.Error("Should find home.Welcome reference")
	}
	if !foundGetUserCount {
		t.Error("Should find home.GetUserCount reference")
	}
	if !foundNonExistent {
		t.Error("Should find home.NonExistentMethod reference (for validation)")
	}
	// Private methods shouldn't be found (regex only matches uppercase methods)
	if foundPrivateMethod {
		t.Error("Should NOT find home.privateMethod reference (private methods not accessible)")
	}
	if !foundIsAuthenticated {
		t.Error("Should find auth.IsAuthenticated reference")
	}
	if !foundCurrentUser {
		t.Error("Should find auth.CurrentUser reference")
	}
	if !foundInvalidMethod {
		t.Error("Should find auth.InvalidMethod reference (for validation)")
	}
}

func TestParseTemplateIncludes(t *testing.T) {
	includes, err := ParseTemplateIncludes("testdata")
	if err != nil {
		t.Fatalf("Failed to parse template includes: %v", err)
	}

	// Should find template includes
	foundValid := false
	foundInvalidPath := false
	foundInvalidRelative := false

	for _, inc := range includes {
		switch inc.TemplateName {
		case "header.html":
			foundValid = true
		case "partials/footer.html":
			foundInvalidPath = true
			if inc.File != "includes.html" {
				t.Errorf("Expected includes.html, got %s", inc.File)
			}
		case "../shared/nav.html":
			foundInvalidRelative = true
		}
	}

	if !foundValid {
		t.Error("Should find valid include (header.html)")
	}
	if !foundInvalidPath {
		t.Error("Should find invalid path include (partials/footer.html)")
	}
	if !foundInvalidRelative {
		t.Error("Should find invalid relative include (../shared/nav.html)")
	}
}

func TestValidateTemplateIncludes(t *testing.T) {
	includes, err := ParseTemplateIncludes("testdata")
	if err != nil {
		t.Fatalf("Failed to parse includes: %v", err)
	}

	errors := ValidateTemplateIncludes(includes)

	// Should have errors for path-based includes
	if len(errors) < 2 {
		t.Errorf("Expected at least 2 errors, got %d", len(errors))
	}

	// Check error messages
	foundPathError := false
	foundRelativeError := false

	for _, err := range errors {
		// Reference is the template name that was included
		if strings.Contains(err.Reference, "partials/footer.html") {
			foundPathError = true
			if err.Problem == "" {
				t.Error("Should have problem description for path error")
			}
			if err.Suggestion == "" {
				t.Error("Should have suggestion for path error")
			}
		}
		if strings.Contains(err.Reference, "../shared/nav.html") {
			foundRelativeError = true
		}
	}

	if !foundPathError {
		t.Error("Should have error for partials/footer.html")
	}
	if !foundRelativeError {
		t.Error("Should have error for ../shared/nav.html")
	}
}
