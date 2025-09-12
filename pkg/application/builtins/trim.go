package builtins

import "strings"

// Trim removes leading and trailing whitespace from a string.
//
// This is a direct wrapper around strings.TrimSpace for template use.
// Removes spaces, tabs, newlines, and other Unicode whitespace.
//
// Parameters:
//   - s: String to trim
//
// Returns:
//   String with leading and trailing whitespace removed
//
// Examples:
//
//	Trim("  Hello World  ")    // "Hello World"
//	Trim("\n\tHello\n\t")      // "Hello"
//	Trim("Hello World")        // "Hello World" (no change)
//	Trim("   ")                // ""
//	Trim("")                   // ""
//	Trim(" Line 1\nLine 2 ")   // "Line 1\nLine 2" (internal whitespace preserved)
//
// Whitespace Definition:
//   Removes all Unicode whitespace characters:
//   - Spaces, tabs, newlines, carriage returns
//   - Non-breaking spaces
//   - Other Unicode space characters
//
// Only Leading/Trailing:
//   Whitespace inside the string is preserved:
//   - "a  b" stays "a  b"
//   - "hello\nworld" stays "hello\nworld"
//
// Template Usage:
//
//	{{ trim .Input }}
//	{{ .Comment | trim }}
//	{{ if eq (trim .Field) "" }}
//	  <!-- Field is empty or only whitespace -->
//	{{ end }}
//
// Common Patterns:
//   - User input cleanup: trim(formInput)
//   - Empty check: trim(s) == ""
//   - Config value cleanup: trim(configValue)
//   - Template output cleanup: trim(rendered)
func Trim(s string) string {
	return strings.TrimSpace(s)
}