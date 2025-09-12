package builtins

import "math"

// Round rounds a number to the nearest integer.
//
// Uses banker's rounding (round half to even), which reduces bias
// in large datasets by rounding .5 to the nearest even number.
//
// Parameters:
//   - n: The number to round
//
// Returns:
//
//	The rounded value as float64
//
// Examples:
//
//	Round(3.7)    // 4
//	Round(3.3)    // 3
//	Round(3.5)    // 4 (rounds up to even)
//	Round(2.5)    // 2 (rounds down to even)
//	Round(4.5)    // 4 (rounds down to even)
//	Round(-2.5)   // -2 (rounds toward even)
//	Round(0)      // 0
//
// Banker's Rounding:
//
//	Also known as "round half to even" or "unbiased rounding":
//	- 0.5 rounds to 0
//	- 1.5 rounds to 2
//	- 2.5 rounds to 2
//	- 3.5 rounds to 4
//
// This reduces statistical bias when summing rounded values.
//
// Template Usage:
//
//	{{ round .Value }}
//	{{ .Percentage | round }}        <!-- round to whole percent -->
//	{{ divf .Total .Count | round }} <!-- round average -->
//
// For Different Rounding:
//   - Use ceil for always rounding up
//   - Use floor for always rounding down
//   - Use custom logic for round-half-up behavior
func Round(n any) float64 {
	return math.Round(toFloat64(n))
}
