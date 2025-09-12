package builtins

import "strings"

// Contains checks if a string contains a substring.
//
// This is a direct wrapper around strings.Contains for template use.
// The check is case-sensitive.
//
// Parameters:
//   - s: String to search in
//   - substr: Substring to search for
//
// Returns:
//   true if s contains substr, false otherwise
//
// Examples:
//
//	Contains("Hello World", "World")  // true
//	Contains("Hello World", "world")  // false (case-sensitive)
//	Contains("Hello World", "o W")    // true
//	Contains("Hello World", "")       // true (empty string is always contained)
//	Contains("", "test")              // false
//	Contains("", "")                  // true
//
// Case Sensitivity:
//   This function is case-sensitive. For case-insensitive search,
//   convert both strings to the same case first:
//   Contains(lower(s), lower(substr))
//
// Template Usage:
//
//	{{ if contains .Description "important" }}
//	  <strong>Important!</strong>
//	{{ end }}
//	
//	{{ if contains .Tags "featured" }}
//	  <span class="featured">Featured</span>
//	{{ end }}
//
// Common Patterns:
//   - Check for keywords: contains text "keyword"
//   - Validate format: contains email "@"
//   - Feature detection: contains userAgent "Mobile"
//   - Tag checking: contains tags "urgent"
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}