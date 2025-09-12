package builtins

// Min returns the smaller of two values.
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
//	The smaller of a and b, preserving type when possible
//
// Examples:
//
//	Min(5, 3)       // 3 (int)
//	Min(3.5, 7.2)   // 3.5 (float64)
//	Min(-5, -3)     // -5 (int)
//	Min(10, 10)     // 10 (int)
//	Min(0, 5)       // 0 (int)
//
// Type Handling:
//   - int, int -> int
//   - float, float -> float64
//   - mixed types -> float64
//   - non-numeric -> compared as float64 (0 for invalid)
//
// Template Usage:
//
//	{{ min .A .B }}
//	{{ min 100 .Value }}             <!-- enforce maximum -->
//	{{ min .MaxSize .RequestedSize }} <!-- cap at maximum -->
//
// Chaining:
//
//	  Can be chained to find minimum of multiple values:
//
//		{{ min .A .B | min .C }}         <!-- min of three values -->
func Min(a, b any) any {
	// Try integer comparison first
	if aInt, aOk := toInt(a); aOk {
		if bInt, bOk := toInt(b); bOk {
			if aInt < bInt {
				return aInt
			}
			return bInt
		}
	}

	// Fall back to float comparison
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat < bFloat {
		return aFloat
	}
	return bFloat
}
