package builtins

import (
	"fmt"
	"math"
)

// FormatPrice formats a numeric value as a USD currency string.
//
// The function formats numbers with exactly 2 decimal places and a dollar sign prefix.
// It properly handles negative values, zero, and very large numbers.
//
// Parameters:
//   - price: The numeric value to format as currency
//
// Returns:
//   A formatted currency string with dollar sign and 2 decimal places
//
// Examples:
//
//	FormatPrice(19.99)     // "$19.99"
//	FormatPrice(19.9)      // "$19.90"
//	FormatPrice(19)        // "$19.00"
//	FormatPrice(-19.99)    // "-$19.99"
//	FormatPrice(0)         // "$0.00"
//	FormatPrice(1000000)   // "$1000000.00"
//
// Type Handling:
//   - int, int32, int64: Converted to float64
//   - float32: Converted to float64
//   - float64: Used directly
//   - Other types: Return "$0.00"
//
// Precision:
//   - Always shows exactly 2 decimal places
//   - Rounds using banker's rounding (round half to even)
//   - No thousands separators (use formatNumber for that)
//
// Template Usage:
//
//	{{ formatPrice .Price }}
//	{{ formatPrice 19.99 }}          <!-- outputs: $19.99 -->
//	{{ .Discount | formatPrice }}    <!-- outputs: -$5.00 -->
func FormatPrice(price interface{}) string {
	var value float64
	
	switch v := price.(type) {
	case float64:
		value = v
	case float32:
		value = float64(v)
	case int:
		value = float64(v)
	case int32:
		value = float64(v)
	case int64:
		value = float64(v)
	default:
		return "$0.00"
	}
	
	// Handle negative values properly
	if value < 0 {
		return fmt.Sprintf("-$%.2f", math.Abs(value))
	}
	
	return fmt.Sprintf("$%.2f", value)
}