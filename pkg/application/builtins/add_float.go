package builtins

// AddFloat performs floating-point addition on two values.
//
// Unlike Add, this always returns float64 regardless of input types.
// Useful when you need to ensure floating-point arithmetic.
//
// Parameters:
//   - a: First addend (converted to float64)
//   - b: Second addend (converted to float64)
//
// Returns:
//   The sum as float64
//
// Examples:
//
//	AddFloat(5, 3)         // 8.0
//	AddFloat(5.5, 2.5)     // 8.0
//	AddFloat("5", 3)       // 8.0 (strings are parsed if numeric)
//	AddFloat(0.1, 0.2)     // 0.30000000000000004 (floating point precision)
//	AddFloat(nil, 5)       // 5.0 (nil treated as 0)
//
// Type Conversion:
//   All types are converted to float64:
//   - Integers: Converted directly
//   - Floats: Used as-is
//   - Strings: Parsed if numeric, otherwise 0
//   - nil: Treated as 0
//   - Other: Converted to 0
//
// Template Usage:
//
//	{{ addf 0.1 0.2 }}              <!-- 0.3 (avoiding integer truncation) -->
//	{{ addf .Percentage 0.5 }}      <!-- add to percentage -->
//	{{ .Rate | addf 0.01 }}         <!-- increment rate -->
//
// Precision Note:
//   Floating point arithmetic may have precision issues.
//   For financial calculations, consider using integer cents.
func AddFloat(a, b interface{}) float64 {
	return toFloat64(a) + toFloat64(b)
}