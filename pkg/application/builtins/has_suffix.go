package builtins

import "strings"

// HasSuffix checks if a string ends with a given suffix.
//
// This is a direct wrapper around strings.HasSuffix for template use.
// The check is case-sensitive.
//
// Parameters:
//   - s: String to check
//   - suffix: Suffix to look for
//
// Returns:
//   true if s ends with suffix, false otherwise
//
// Examples:
//
//	HasSuffix("Hello World", "World")   // true
//	HasSuffix("Hello World", "world")   // false (case-sensitive)
//	HasSuffix("Hello World", "Hello")   // false
//	HasSuffix("Hello World", "")        // true (empty suffix always matches)
//	HasSuffix("", "test")               // false
//	HasSuffix("", "")                   // true
//	HasSuffix("image.png", ".png")      // true
//
// Template Usage:
//
//	{{ if hasSuffix .Filename ".pdf" }}
//	  <i class="icon-pdf"></i>
//	{{ end }}
//	
//	{{ if hasSuffix .Email ".edu" }}
//	  <span class="academic">Academic Email</span>
//	{{ end }}
//	
//	{{ if hasSuffix .URL "/" }}
//	  <!-- URL ends with slash -->
//	{{ else }}
//	  <!-- URL doesn't end with slash -->
//	{{ end }}
//
// Common Patterns:
//   - File extension check: hasSuffix filename ".jpg"
//   - Domain validation: hasSuffix email "@company.com"
//   - Path normalization: hasSuffix path "/"
//   - String ending check: hasSuffix text "..."
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}