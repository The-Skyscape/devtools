package builtins

// DivideFloat performs floating-point division on two values.
//
// Always returns float64. Returns 0 for division by zero.
//
// Parameters:
//   - a: Dividend (converted to float64)
//   - b: Divisor (converted to float64)
//
// Returns:
//   The quotient as float64, or 0 if b is zero
//
// Examples:
//
//	DivideFloat(10, 3)       // 3.333333...
//	DivideFloat(10, 2)       // 5.0
//	DivideFloat("100", 4)    // 25.0 (strings parsed if numeric)
//	DivideFloat(15, 0)       // 0.0 (safe division by zero)
//	DivideFloat(nil, 5)      // 0.0 (nil treated as 0)
//
// Template Usage:
//
//	{{ divf 10 3 }}                  <!-- 3.333... -->
//	{{ divf .Total .Count }}         <!-- calculate average -->
//	{{ .Percentage | divf 100 }}     <!-- convert from percentage -->
//
// Division by Zero:
//   Returns 0 instead of +Inf, -Inf, or NaN.
//   This makes templates safer but may hide errors.
//
// Common Patterns:
//   - Calculate average: divf total count
//   - Convert from percentage: divf value 100
//   - Calculate rate: divf completed total
func DivideFloat(a, b interface{}) float64 {
	divisor := toFloat64(b)
	if divisor == 0 {
		return 0
	}
	return toFloat64(a) / divisor
}