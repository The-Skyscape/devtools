package builtins

// Coalesce returns the first non-empty value from a list.
//
// Iterates through the provided values and returns the first one
// that is not empty (non-zero for its type).
//
// Parameters:
//   - values: Variadic list of values to check
//
// Returns:
//   First non-empty value, or nil if all are empty
//
// Examples:
//
//	Coalesce("", "default", "backup")  // "default"
//	Coalesce(0, 0, 42, 100)            // 42
//	Coalesce(nil, nil, "value")        // "value"
//	Coalesce("first", "second")        // "first"
//	Coalesce()                         // nil
//	Coalesce("", "", "")               // nil
//	Coalesce(false, true, false)       // true
//
// Empty Detection:
//   Uses the same empty detection as Default():
//   - Strings: Empty string ""
//   - Numbers: Zero value (0, 0.0)
//   - Booleans: false
//   - Slices/Maps: nil or length 0
//   - Pointers/Interfaces: nil
//
// Template Usage:
//
//	{{ coalesce .Nickname .Username .Email }}
//	{{ coalesce .CustomTitle .DefaultTitle "Untitled" }}
//	{{ coalesce .Port 8080 }}
//
// Common Patterns:
//   - Fallback chains: coalesce(primary, secondary, fallback)
//   - Config hierarchy: coalesce(envVar, configFile, default)
//   - Display values: coalesce(displayName, username, "Guest")
//   - Optional fields: coalesce(optional, required)
//
// Difference from Default:
//   - Default: Two values (value and default)
//   - Coalesce: Multiple values (returns first non-empty)
//
// Note: Particularly useful for template logic where
// multiple fallback values need to be checked.
func Coalesce(values ...interface{}) interface{} {
	for _, v := range values {
		if !IsEmpty(v) {
			return v
		}
	}
	return nil
}