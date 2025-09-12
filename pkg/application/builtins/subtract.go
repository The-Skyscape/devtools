package builtins

// Subtract performs subtraction on two numeric values with type safety.
//
// Like Add, this function preserves integer types when both operands are integers.
//
// Parameters:
//   - a: Minuend (number to subtract from)
//   - b: Subtrahend (number to subtract)
//
// Returns:
//
//	The difference a - b, preserving integer type when possible
//
// Examples:
//
//	Subtract(10, 3)      // 7 (int)
//	Subtract(5.5, 2.5)   // 3.0 (float64)
//	Subtract(10, 15)     // -5 (int)
//	Subtract(0, 5)       // -5 (int)
//
// Template Usage:
//
//	{{ sub 10 3 }}                    <!-- 7 -->
//	{{ sub .Total .Discount }}        <!-- discounted total -->
//	{{ .Count | sub 1 }}              <!-- decrements count -->
func Subtract(a, b any) any {
	// Try integer subtraction first
	if aInt, aOk := toInt(a); aOk {
		if bInt, bOk := toInt(b); bOk {
			return aInt - bInt
		}
	}

	// Fall back to float subtraction
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat == 0 && bFloat == 0 && !isNumeric(a) && !isNumeric(b) {
		return 0
	}
	return aFloat - bFloat
}
