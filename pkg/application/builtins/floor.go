package builtins

import "math"

// Floor returns the largest integer less than or equal to the input.
//
// This function always rounds down toward negative infinity.
//
// Parameters:
//   - n: The number to floor
//
// Returns:
//   The floored value as float64
//
// Examples:
//
//	Floor(3.7)    // 3
//	Floor(3.0)    // 3
//	Floor(-2.3)   // -3 (rounds down, not toward zero)
//	Floor(-2.0)   // -2
//	Floor(0.9)    // 0
//	Floor(5)      // 5
//
// Negative Numbers:
//   Note that floor rounds toward negative infinity, not toward zero:
//   - Floor(2.5) = 2
//   - Floor(-2.5) = -3 (not -2)
//
// Template Usage:
//
//	{{ floor .Value }}
//	{{ floor .Average }}             <!-- round down to integer -->
//	{{ divf .Total 100 | floor }}    <!-- pages needed (round down) -->
//
// Common Uses:
//   - Converting floats to integers (always rounding down)
//   - Calculating complete units (e.g., complete pages, batches)
//   - Array indexing after division
func Floor(n interface{}) float64 {
	return math.Floor(toFloat64(n))
}