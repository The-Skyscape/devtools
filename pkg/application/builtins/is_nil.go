package builtins

import "reflect"

// IsNil checks if a value is nil.
//
// Safely checks if a value is nil, handling both typed
// and untyped nil values.
//
// Parameters:
//   - v: Value to check for nil
//
// Returns:
//
//	true if the value is nil, false otherwise
//
// Examples:
//
//	IsNil(nil)                  // true
//	IsNil("hello")              // false
//	IsNil(0)                    // false
//	IsNil("")                   // false
//	var p *int
//	IsNil(p)                    // true
//	var s []int
//	IsNil(s)                    // true
//	IsNil([]int{})              // false
//	var m map[string]int
//	IsNil(m)                    // true
//	IsNil(map[string]int{})     // false
//
// Nil-able Types:
//
//	Only these types can be nil:
//	- Pointers
//	- Slices
//	- Maps
//	- Channels
//	- Interfaces
//	- Functions
//
// Non-nil Types:
//
//	These types can never be nil:
//	- Basic types (int, string, bool, etc.)
//	- Structs
//	- Arrays (not slices)
//
// Template Usage:
//
//	{{ if isNil .OptionalField }}
//	  <p>Field not set</p>
//	{{ else }}
//	  <p>{{ .OptionalField }}</p>
//	{{ end }}
//
//	{{ if not (isNil .Data) }}
//	  <!-- Process data -->
//	{{ end }}
//
// Common Patterns:
//   - Null checks: if isNil(ptr) { return error }
//   - Optional fields: isNil(optional) ? skip : process
//   - Safety checks: !isNil(v) && v.IsValid()
//   - Template guards: {{ if not (isNil .Field) }}...{{ end }}
//
// Difference from IsEmpty:
//   - IsNil: Only checks for nil
//   - IsEmpty: Checks for zero values (includes nil)
//     Example: IsNil("") = false, IsEmpty("") = true
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}
