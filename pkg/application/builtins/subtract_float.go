package builtins

// SubtractFloat performs floating-point subtraction on two values.
//
// Always returns float64 regardless of input types.
//
// Parameters:
//   - a: Minuend (converted to float64)
//   - b: Subtrahend (converted to float64)
//
// Returns:
//
//	The difference as float64
//
// Examples:
//
//	SubtractFloat(10.5, 3.2)  // 7.3
//	SubtractFloat(10, 3)      // 7.0
//	SubtractFloat("10", 3)    // 7.0 (strings parsed if numeric)
//	SubtractFloat(5, 10)      // -5.0
//	SubtractFloat(nil, 5)     // -5.0 (nil treated as 0)
//
// Template Usage:
//
//	{{ subf 10.5 3.2 }}              <!-- 7.3 -->
//	{{ subf .Total .Discount }}      <!-- apply discount -->
//	{{ .Price | subf .Reduction }}   <!-- reduce price -->
//
// Precision Note:
//
//	Subject to floating point precision limitations.
//	For exact decimal arithmetic, consider scaled integers.
func SubtractFloat(a, b any) float64 {
	return toFloat64(a) - toFloat64(b)
}
