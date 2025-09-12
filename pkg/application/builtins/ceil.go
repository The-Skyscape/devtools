package builtins

import "math"

// Ceil returns the smallest integer greater than or equal to the input.
//
// This function always rounds up toward positive infinity.
//
// Parameters:
//   - n: The number to ceil
//
// Returns:
//   The ceiled value as float64
//
// Examples:
//
//	Ceil(3.3)     // 4
//	Ceil(3.0)     // 3
//	Ceil(-2.7)    // -2 (rounds up, not away from zero)
//	Ceil(-2.0)    // -2
//	Ceil(0.1)     // 1
//	Ceil(5)       // 5
//
// Negative Numbers:
//   Note that ceil rounds toward positive infinity, not away from zero:
//   - Ceil(2.5) = 3
//   - Ceil(-2.5) = -2 (not -3)
//
// Template Usage:
//
//	{{ ceil .Value }}
//	{{ ceil .Average }}              <!-- round up to integer -->
//	{{ divf .Total .PageSize | ceil }} <!-- pages needed (round up) -->
//
// Common Uses:
//   - Calculating required capacity (always round up)
//   - Pagination (number of pages needed)
//   - Resource allocation (minimum units needed)
func Ceil(n interface{}) float64 {
	return math.Ceil(toFloat64(n))
}