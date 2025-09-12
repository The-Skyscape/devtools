package builtins

// Add performs addition on two numeric values with type safety.
//
// The function attempts to preserve integer types when both operands are integers,
// returning float64 only when necessary (mixed types or float operands).
//
// Parameters:
//   - a: First operand (int or float64)
//   - b: Second operand (int or float64)
//
// Returns:
//
//	The sum of a and b, preserving integer type when possible
//
// Type Handling:
//   - int + int = int
//   - float64 + float64 = float64
//   - int + float64 = float64
//   - Other types = 0
//
// Examples:
//
//	Add(5, 3)         // 8 (int)
//	Add(5.5, 2.5)     // 8.0 (float64)
//	Add(5, 2.5)       // 7.5 (float64)
//	Add(100, -50)     // 50 (int)
//	Add("5", 3)       // 0 (invalid type)
//
// Template Usage:
//
//	{{ add 10 5 }}                    <!-- 15 -->
//	{{ add .Count 1 }}                <!-- increments count -->
//	{{ add .Price .Tax }}             <!-- total with tax -->
//
// Chaining:
//
//	  Can be chained with other math functions:
//
//		{{ add 10 5 | add 3 }}           <!-- 18 -->
//		{{ .Base | add .Bonus | add .Extra }}
func Add(a, b any) any {
	// Try integer addition first
	if aInt, aOk := toInt(a); aOk {
		if bInt, bOk := toInt(b); bOk {
			return aInt + bInt
		}
	}

	// Fall back to float addition
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat == 0 && bFloat == 0 && !isNumeric(a) && !isNumeric(b) {
		return 0
	}
	return aFloat + bFloat
}
