package builtins

import "strings"

// Title converts a string to title case.
//
// This is a direct wrapper around strings.Title for template use.
// Capitalizes the first letter of each word.
//
// Parameters:
//   - s: String to convert
//
// Returns:
//   Title-cased version of the string
//
// Examples:
//
//	Title("hello world")      // "Hello World"
//	Title("the quick brown")  // "The Quick Brown"
//	Title("UPPERCASE TEXT")   // "UPPERCASE TEXT" (doesn't lowercase)
//	Title("mixedCase text")   // "MixedCase Text"
//	Title("")                // ""
//	Title("john's bike")     // "John'S Bike" (apostrophe creates new word)
//	Title("html-parser")     // "Html-Parser" (hyphen creates new word)
//
// Word Boundaries:
//   A word is defined as a sequence of letters. Any non-letter
//   character creates a word boundary:
//   - Spaces, punctuation, numbers all create boundaries
//   - "it's" becomes "It'S" (two words)
//   - "co-op" becomes "Co-Op" (two words)
//
// Limitations:
//   - Doesn't handle articles properly ("The", "A", "An")
//   - Doesn't lowercase existing uppercase letters
//   - Simple algorithm, not linguistically aware
//
// Template Usage:
//
//	{{ title .Name }}
//	{{ .Description | title }}
//	<h2>{{ title .Category }}</h2>
//
// Common Patterns:
//   - Name formatting: title(firstName + " " + lastName)
//   - Headers: title(sectionName)
//   - Menu items: title(menuLabel)
//
// Note: strings.Title is deprecated in Go 1.18+. Consider using
// cases.Title from golang.org/x/text/cases for better Unicode support.
func Title(s string) string {
	return strings.Title(s)
}