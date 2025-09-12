package builtins

// Divide performs division on two numeric values with zero-division protection.
//
// Integer division is performed when both operands are integers and the result
// is a whole number. Otherwise, floating-point division is used.
//
// Parameters:
//   - a: Dividend (number to be divided)
//   - b: Divisor (number to divide by)
//
// Returns:
//
//	The quotient a / b, or 0 if b is zero
//
// Division by Zero:
//
//	Returns 0 instead of infinity or panic
//
// Examples:
//
//	Divide(10, 2)        // 5 (int)
//	Divide(10, 3)        // 3.333... (float64)
//	Divide(10, 0)        // 0 (safe division by zero)
//	Divide(10.5, 2)      // 5.25 (float64)
//	Divide(9, 3)         // 3 (int - exact division)
//
// Template Usage:
//
//	{{ div 10 2 }}                    <!-- 5 -->
//	{{ div .Total .Count }}           <!-- average -->
//	{{ .Percentage | div 100 }}       <!-- convert from percentage -->
//
// Integer Division Note:
//
//	When both operands are integers and you want integer division
//	(truncating the remainder), the result will be integer only
//	if it divides evenly. For forced integer division, cast the
//	result or use a dedicated integer division function.
func Divide(a, b any) any {
	// Check for division by zero early
	bFloat := toFloat64(b)
	if bFloat == 0 {
		return 0
	}

	// Try integer division if both are integers and divide evenly
	if aInt, aOk := toInt(a); aOk {
		if bInt, bOk := toInt(b); bOk && bInt != 0 {
			if aInt%bInt == 0 {
				return aInt / bInt
			}
			// Doesn't divide evenly, use float
			return float64(aInt) / float64(bInt)
		}
	}

	// Fall back to float division
	aFloat := toFloat64(a)
	if aFloat == 0 && !isNumeric(a) {
		return 0
	}
	return aFloat / bFloat
}
