package builtins

// Max returns the larger of two values.
//
// The function preserves integer type when both inputs are integers,
// otherwise returns float64.
//
// Parameters:
//   - a: First value to compare
//   - b: Second value to compare
//
// Returns:
//
//	The larger of a and b, preserving type when possible
//
// Examples:
//
//	Max(5, 3)       // 5 (int)
//	Max(3.5, 7.2)   // 7.2 (float64)
//	Max(-5, -3)     // -3 (int)
//	Max(10, 10)     // 10 (int)
//	Max(0, -5)      // 0 (int)
//
// Type Handling:
//   - int, int -> int
//   - float, float -> float64
//   - mixed types -> float64
//   - non-numeric -> compared as float64 (0 for invalid)
//
// Template Usage:
//
//	{{ max .A .B }}
//	{{ max 0 .Value }}              <!-- ensure non-negative -->
//	{{ max .MinSize .RequestedSize }} <!-- enforce minimum -->
//
// Chaining:
//
//	  Can be chained to find maximum of multiple values:
//
//		{{ max .A .B | max .C }}        <!-- max of three values -->
func Max(a, b any) any {
	// Try integer comparison first
	if aInt, aOk := toInt(a); aOk {
		if bInt, bOk := toInt(b); bOk {
			if aInt > bInt {
				return aInt
			}
			return bInt
		}
	}

	// Fall back to float comparison
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat > bFloat {
		return aFloat
	}
	return bFloat
}
