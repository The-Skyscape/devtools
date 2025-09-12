package builtins

// Multiply performs multiplication on two numeric values.
//
// The function handles integer overflow by automatically promoting to float64
// when the result would exceed integer bounds.
//
// Parameters:
//   - a: First factor
//   - b: Second factor
//
// Returns:
//   The product a * b, as int if possible, otherwise float64
//
// Examples:
//
//	Multiply(5, 3)       // 15 (int)
//	Multiply(2.5, 4)     // 10.0 (float64)
//	Multiply(100, 0)     // 0 (int)
//	Multiply(-5, 3)      // -15 (int)
//
// Template Usage:
//
//	{{ mul 5 3 }}                     <!-- 15 -->
//	{{ mul .Quantity .Price }}        <!-- total price -->
//	{{ .Rate | mul 100 }}             <!-- convert to percentage -->
func Multiply(a, b interface{}) interface{} {
	// Try integer multiplication first
	if aInt, aOk := toInt(a); aOk {
		if bInt, bOk := toInt(b); bOk {
			result := aInt * bInt
			// Check for overflow by dividing back
			if bInt != 0 && result/bInt == aInt {
				return result
			}
			// Overflow occurred, use float
			return float64(aInt) * float64(bInt)
		}
	}
	
	// Fall back to float multiplication
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat == 0 && bFloat == 0 && !isNumeric(a) && !isNumeric(b) {
		return 0
	}
	return aFloat * bFloat
}