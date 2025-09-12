package builtins

// MultiplyFloat performs floating-point multiplication on two values.
//
// Always returns float64 regardless of input types.
//
// Parameters:
//   - a: First factor (converted to float64)
//   - b: Second factor (converted to float64)
//
// Returns:
//
//	The product as float64
//
// Examples:
//
//	MultiplyFloat(2.5, 4)     // 10.0
//	MultiplyFloat(0.1, 10)    // 1.0
//	MultiplyFloat("5", 3)     // 15.0 (strings parsed if numeric)
//	MultiplyFloat(0.15, 100)  // 15.0 (convert to percentage)
//	MultiplyFloat(nil, 5)     // 0.0 (nil treated as 0)
//
// Template Usage:
//
//	{{ mulf .Rate 100 }}             <!-- convert rate to percentage -->
//	{{ mulf .Price .Quantity }}      <!-- calculate total -->
//	{{ .Tax | mulf .Subtotal }}      <!-- calculate tax amount -->
//
// Common Patterns:
//   - Convert decimal to percentage: mulf value 100
//   - Calculate percentage of value: mulf value 0.15
//   - Scale values: mulf value scaleFactor
func MultiplyFloat(a, b any) float64 {
	return toFloat64(a) * toFloat64(b)
}
