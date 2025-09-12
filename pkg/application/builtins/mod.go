package builtins

import "math"

// Mod returns the remainder of integer division (modulo operation).
//
// For integer inputs, performs integer modulo. For floats, uses math.Mod.
//
// Parameters:
//   - a: Dividend
//   - b: Divisor
//
// Returns:
//   The remainder of a / b, or 0 if b is zero
//
// Examples:
//
//	Mod(10, 3)    // 1
//	Mod(10, 5)    // 0
//	Mod(7, 2)     // 1
//	Mod(-7, 3)    // -1 (follows Go's % operator semantics)
//
// Template Usage:
//
//	{{ mod .Index 2 }}  <!-- for even/odd checking -->
//	{{ if eq (mod .Index 2) 0 }}even{{ else }}odd{{ end }}
//
// Negative Values:
//   Follows Go's modulo semantics where the sign matches the dividend:
//   - Mod(7, 3) = 1
//   - Mod(-7, 3) = -1
//   - Mod(7, -3) = 1
//   - Mod(-7, -3) = -1
func Mod(a, b interface{}) interface{} {
	// Try integer modulo first
	if aInt, aOk := toInt(a); aOk {
		if bInt, bOk := toInt(b); bOk && bInt != 0 {
			return aInt % bInt
		}
	}
	
	// Fall back to float modulo
	bFloat := toFloat64(b)
	if bFloat == 0 {
		return 0
	}
	return math.Mod(toFloat64(a), bFloat)
}