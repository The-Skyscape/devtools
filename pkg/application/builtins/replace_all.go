package builtins

import "strings"

// ReplaceAll replaces all occurrences of a substring in a string.
//
// This is a direct wrapper around strings.ReplaceAll for template use.
// Replaces every occurrence of old with new in the string.
//
// Parameters:
//   - s: String to search in
//   - old: Substring to find
//   - new: Replacement string
//
// Returns:
//   String with all occurrences replaced
//
// Examples:
//
//	ReplaceAll("Hello World", "o", "0")     // "Hell0 W0rld"
//	ReplaceAll("foo bar foo", "foo", "baz") // "baz bar baz"
//	ReplaceAll("test", "x", "y")            // "test" (no match)
//	ReplaceAll("", "a", "b")                // ""
//	ReplaceAll("aaa", "a", "")              // "" (removes all)
//	ReplaceAll("path/to/file", "/", "-")   // "path-to-file"
//
// Edge Cases:
//   - Empty old string returns original (Go behavior)
//   - Empty new string removes all occurrences
//   - No matches returns original string
//   - Overlapping matches handled correctly
//
// Template Usage:
//
//	{{ replaceAll .Path "/" "-" }}
//	{{ .Content | replaceAll "\n" "<br>" }}
//	{{ replaceAll .Text " " "_" }}
//
// Common Patterns:
//   - Path normalization: replaceAll path "\\" "/"
//   - HTML escaping: replaceAll text "<" "&lt;"
//   - Slug creation: replaceAll title " " "-"
//   - Remove characters: replaceAll s char ""
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}