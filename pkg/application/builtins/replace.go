package builtins

import "strings"

// Replace replaces all occurrences of a substring with another string.
//
// This is a direct wrapper around strings.ReplaceAll for template use.
// The replacement is case-sensitive.
//
// Parameters:
//   - s: String to search in
//   - old: Substring to replace
//   - new: Replacement string
//
// Returns:
//   String with all occurrences of old replaced by new
//
// Examples:
//
//	Replace("Hello World", "o", "0")         // "Hell0 W0rld"
//	Replace("Hello World", "World", "Go")    // "Hello Go"
//	Replace("Hello World", "world", "Go")    // "Hello World" (case-sensitive)
//	Replace("aaaa", "aa", "b")              // "bb"
//	Replace("Hello", "x", "y")              // "Hello" (no match)
//	Replace("Hello", "", "x")               // "xHxexlxlxox" (empty string matches everywhere)
//	Replace("", "a", "b")                   // ""
//
// Special Cases:
//   - Empty old string: Inserts new between every character
//   - Empty new string: Removes all occurrences of old
//   - No matches: Returns original string unchanged
//
// Template Usage:
//
//	{{ replace .Text "\\n" "<br>" }}        <!-- Convert newlines to HTML breaks -->
//	{{ replace .URL " " "%20" }}            <!-- URL encode spaces -->
//	{{ replace .Template "{{name}}" .Name }} <!-- Simple template substitution -->
//	
// Common Patterns:
//   - HTML escaping: replace text "<" "&lt;"
//   - Path normalization: replace path "\\" "/"
//   - Remove characters: replace text "unwanted" ""
//   - Simple templating: replace template "{{var}}" value
//
// Note:
//   This replaces ALL occurrences. For single replacement or
//   limited replacements, a different function would be needed.
func Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}