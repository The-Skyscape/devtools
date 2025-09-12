package builtins

import "strings"

// Split divides a string into a slice using a separator.
//
// This is a direct wrapper around strings.Split for template use.
//
// Parameters:
//   - s: String to split
//   - sep: Separator to split on
//
// Returns:
//   Slice of strings split by separator
//
// Examples:
//
//	Split("a,b,c", ",")           // []string{"a", "b", "c"}
//	Split("path/to/file", "/")    // []string{"path", "to", "file"}
//	Split("hello world", " ")     // []string{"hello", "world"}
//	Split("no-separator", ",")    // []string{"no-separator"}
//	Split("", ",")                // []string{""}
//	Split("a,,c", ",")            // []string{"a", "", "c"} (empty strings preserved)
//
// Special Cases:
//   - Empty string returns slice with one empty string
//   - No separator found returns slice with original string
//   - Empty separator splits into individual characters
//   - Consecutive separators create empty strings in result
//
// Template Usage:
//
//	{{ split .CSV "," }}
//	{{ split .Path "/" }}
//	{{ range split .Tags "," }}
//	  <span class="tag">{{ . }}</span>
//	{{ end }}
//
// Common Patterns:
//   - Parse CSV: split line ","
//   - Extract path segments: split path "/"
//   - Parse tags: split tags ","
//   - Word splitting: split text " "
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}