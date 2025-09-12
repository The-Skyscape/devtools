package builtins

import "strings"

// Lower converts a string to lowercase.
//
// This is a direct wrapper around strings.ToLower for template use.
// Handles Unicode characters correctly.
//
// Parameters:
//   - s: String to convert
//
// Returns:
//   Lowercase version of the string
//
// Examples:
//
//	Lower("Hello World")    // "hello world"
//	Lower("UPPERCASE")      // "uppercase"
//	Lower("MixedCase123")   // "mixedcase123"
//	Lower("already lower")  // "already lower"
//	Lower("")              // ""
//	Lower("Café")          // "café" (Unicode aware)
//	Lower("ἈΒΓΔ")          // "ἀβγδ" (Greek letters)
//
// Unicode Support:
//   Correctly handles Unicode characters according to Unicode
//   case mapping rules. Works with all scripts that have case.
//
// Template Usage:
//
//	{{ lower .Title }}
//	{{ .Email | lower }}
//	{{ if eq (lower .Type) "admin" }}
//	  <!-- Case-insensitive comparison -->
//	{{ end }}
//
// Common Patterns:
//   - Normalize for comparison: lower(a) == lower(b)
//   - Email normalization: lower(email)
//   - CSS classes: lower(className)
//   - Username normalization: lower(username)
func Lower(s string) string {
	return strings.ToLower(s)
}