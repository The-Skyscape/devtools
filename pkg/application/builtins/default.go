package builtins

// Default returns a default value if the input is empty.
//
// Checks if the value is empty (zero value for its type) and
// returns the default if true, otherwise returns the original value.
//
// Parameters:
//   - v: Value to check
//   - def: Default value to return if v is empty
//
// Returns:
//   Original value if non-empty, default value otherwise
//
// Examples:
//
//	Default("", "unnamed")         // "unnamed"
//	Default("John", "unnamed")     // "John"
//	Default(0, 100)                // 100
//	Default(42, 100)               // 42
//	Default(nil, "default")        // "default"
//	Default([]int{}, []int{1, 2}) // []int{1, 2}
//	Default([]int{3}, []int{1, 2}) // []int{3}
//
// Empty Detection:
//   - Strings: Empty string ""
//   - Numbers: Zero value (0, 0.0)
//   - Booleans: false
//   - Slices/Maps: nil or length 0
//   - Pointers: nil
//   - Interfaces: nil
//
// Template Usage:
//
//	{{ default .Name "Anonymous" }}
//	{{ .Count | default 1 }}
//	{{ default .Title "Untitled" }}
//
// Common Patterns:
//   - Form fields: default(input, placeholder)
//   - Config values: default(config, defaultValue)
//   - Display names: default(username, "Guest")
//   - Pagination: default(page, 1)
//
// Type Safety:
//   Both parameters can be any type. The function
//   returns the same type as provided, maintaining
//   type consistency in templates.
func Default(v, def interface{}) interface{} {
	if IsEmpty(v) {
		return def
	}
	return v
}