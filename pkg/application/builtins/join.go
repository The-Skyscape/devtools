package builtins

import "strings"

// Join concatenates string slice elements with a separator.
//
// This is a direct wrapper around strings.Join for template use.
//
// Parameters:
//   - items: Slice of strings to join
//   - sep: Separator string to place between elements
//
// Returns:
//   Joined string with separator between elements
//
// Examples:
//
//	Join([]string{"a", "b", "c"}, ", ")     // "a, b, c"
//	Join([]string{"path", "to", "file"}, "/") // "path/to/file"
//	Join([]string{"one"}, ", ")             // "one"
//	Join([]string{}, ", ")                  // ""
//	Join(nil, ", ")                          // ""
//	Join([]string{"a", "", "c"}, "-")       // "a--c" (empty strings included)
//
// Edge Cases:
//   - Empty slice returns empty string
//   - Single element returns that element (no separator)
//   - Empty strings in slice are preserved
//   - nil slice returns empty string
//
// Template Usage:
//
//	{{ join .Tags ", " }}
//	{{ join .Path "/" }}
//	{{ .Authors | join " and " }}
//
// Common Patterns:
//   - CSV format: join items ", "
//   - Path building: join segments "/"
//   - Natural language lists: join items ", " for all but last, then " and "
//   - HTML classes: join classes " "
func Join(items []string, sep string) string {
	if items == nil {
		return ""
	}
	return strings.Join(items, sep)
}