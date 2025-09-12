package builtins

import "fmt"

// FormatPercent formats a numeric value as a percentage string with one decimal place.
//
// The function expects values in percentage form (e.g., 50 for 50%, not 0.5).
// To convert decimal values, multiply by 100 before calling this function.
//
// Parameters:
//   - value: The percentage value to format (0-100 scale)
//
// Returns:
//   A formatted percentage string with % suffix and 1 decimal place
//
// Examples:
//
//	FormatPercent(50)      // "50.0%"
//	FormatPercent(50.5)    // "50.5%"
//	FormatPercent(100)     // "100.0%"
//	FormatPercent(0)       // "0.0%"
//	FormatPercent(-10)     // "-10.0%"
//	FormatPercent(33.333)  // "33.3%"
//
// Type Handling:
//   - int, int32, int64: Converted to float64
//   - float32: Converted to float64
//   - float64: Used directly
//   - Other types: Return "0.0%"
//
// Precision:
//   - Always shows exactly 1 decimal place
//   - Rounds to nearest tenth
//   - Handles negative percentages
//
// Common Patterns:
//
//	// Converting from decimal to percentage:
//	{{ mulf .Rate 100 | formatPercent }}  // 0.15 -> "15.0%"
//	
//	// Direct percentage values:
//	{{ formatPercent .SuccessRate }}      // 95.5 -> "95.5%"
//
// Template Usage:
//
//	{{ formatPercent 75.5 }}               <!-- outputs: 75.5% -->
//	{{ .GrowthRate | formatPercent }}      <!-- outputs: 12.3% -->
func FormatPercent(value any) string {
	var percent float64
	
	switch v := value.(type) {
	case float64:
		percent = v
	case float32:
		percent = float64(v)
	case int:
		percent = float64(v)
	case int32:
		percent = float64(v)
	case int64:
		percent = float64(v)
	default:
		return "0.0%"
	}
	
	return fmt.Sprintf("%.1f%%", percent)
}