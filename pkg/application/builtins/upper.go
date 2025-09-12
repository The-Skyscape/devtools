package builtins

import "strings"

// Upper converts a string to uppercase.
//
// This is a direct wrapper around strings.ToUpper for template use.
// Handles Unicode characters correctly.
//
// Parameters:
//   - s: String to convert
//
// Returns:
//   Uppercase version of the string
//
// Examples:
//
//	Upper("Hello World")    // "HELLO WORLD"
//	Upper("lowercase")      // "LOWERCASE"
//	Upper("MixedCase123")   // "MIXEDCASE123"
//	Upper("ALREADY UPPER")  // "ALREADY UPPER"
//	Upper("")              // ""
//	Upper("café")          // "CAFÉ" (Unicode aware)
//	Upper("αβγδ")          // "ΑΒΓΔ" (Greek letters)
//
// Unicode Support:
//   Correctly handles Unicode characters according to Unicode
//   case mapping rules. Works with all scripts that have case.
//
// Template Usage:
//
//	{{ upper .Code }}
//	{{ .Status | upper }}
//	<h1>{{ upper .Title }}</h1>
//
// Common Patterns:
//   - Emphasis: upper(importantText)
//   - Constants/codes: upper(code)
//   - Headers: upper(heading)
//   - Status indicators: upper(status)
func Upper(s string) string {
	return strings.ToUpper(s)
}