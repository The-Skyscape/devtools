package builtins

import "reflect"

// IsEmpty checks if a value is empty (zero value for its type).
//
// Determines if a value is considered "empty" based on Go's
// definition of zero values.
//
// Parameters:
//   - v: Value to check for emptiness
//
// Returns:
//   true if the value is empty, false otherwise
//
// Examples:
//
//	IsEmpty("")              // true
//	IsEmpty("hello")         // false
//	IsEmpty(0)               // true
//	IsEmpty(42)              // false
//	IsEmpty(false)           // true
//	IsEmpty(true)            // false
//	IsEmpty(nil)             // true
//	IsEmpty([]int{})         // true
//	IsEmpty([]int{1})        // false
//	IsEmpty(map[string]int{}) // true
//
// Empty Values by Type:
//   - nil: Always empty
//   - bool: false is empty
//   - int, float: 0 is empty
//   - string: "" is empty
//   - slice, map: nil or length 0 is empty
//   - pointer: nil is empty
//   - interface: nil is empty
//   - struct: Never empty (even zero-value struct)
//
// Template Usage:
//
//	{{ if isEmpty .Field }}
//	  <p>Field is empty</p>
//	{{ end }}
//	
//	{{ if not (isEmpty .Items) }}
//	  {{ range .Items }}...{{ end }}
//	{{ end }}
//
// Common Patterns:
//   - Validation: if isEmpty(input) { return error }
//   - Conditionals: isEmpty(list) ? showEmpty : showList
//   - Filtering: skip if isEmpty(value)
//   - Defaults: isEmpty(v) ? defaultValue : v
//
// Note: Struct values are never considered empty,
// even if all fields are zero values. Use specific
// field checks for struct emptiness.
func IsEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.String:
		return rv.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	default:
		// Structs and other types are never empty
		return false
	}
}