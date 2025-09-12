package builtins

import "fmt"

// ToString converts any value to its string representation.
//
// Converts any Go value to a string using fmt.Sprint.
// Handles all types including nil values.
//
// Parameters:
//   - v: Any value to convert
//
// Returns:
//   String representation of the value
//
// Examples:
//
//	ToString(42)              // "42"
//	ToString(3.14)            // "3.14"
//	ToString(true)            // "true"
//	ToString("hello")         // "hello"
//	ToString(nil)             // "<nil>"
//	ToString([]int{1, 2, 3})  // "[1 2 3]"
//	ToString(struct{}{})      // "{}"
//
// Type Conversions:
//   - Numbers: Decimal representation
//   - Booleans: "true" or "false"
//   - Strings: Unchanged
//   - nil: "<nil>"
//   - Arrays/Slices: Space-separated values in brackets
//   - Structs: Field values in braces
//   - Pointers: Dereferenced value or "<nil>"
//
// Template Usage:
//
//	{{ toString .Value }}
//	{{ .Number | toString }}
//	{{ if eq (toString .Field) "" }}
//	  <!-- Field is empty -->
//	{{ end }}
//
// Common Patterns:
//   - Safe string conversion: toString(anyValue)
//   - Debug output: toString(complexStruct)
//   - String concatenation: toString(a) + toString(b)
//   - Comparison: toString(a) == toString(b)
//
// Note: Uses fmt.Sprint internally, which provides
// consistent formatting across all Go types.
func ToString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprint(v)
}