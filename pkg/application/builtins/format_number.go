package builtins

import (
	"fmt"
	"math"
	"strings"
)

// FormatNumber formats a numeric value with thousands separators.
//
// The function adds commas as thousands separators and optionally shows
// up to 2 decimal places (trailing zeros are removed).
//
// Parameters:
//   - n: The number to format (int, int32, int64, float32, float64)
//
// Returns:
//   A formatted number string with commas, or the original value as string if not numeric
//
// Examples:
//
//	FormatNumber(1000)        // "1,000"
//	FormatNumber(1000000)     // "1,000,000"
//	FormatNumber(1234.56)     // "1,234.56"
//	FormatNumber(1234.50)     // "1,234.5"
//	FormatNumber(1234.00)     // "1,234"
//	FormatNumber(-5000)       // "-5,000"
//	FormatNumber(0)           // "0"
//	FormatNumber(999)         // "999"
//
// Type Handling:
//   - int, int32, int64: Formatted as integers
//   - float32, float64: Formatted with up to 2 decimal places
//   - Other types: Converted to string using fmt.Sprint
//
// Decimal Handling:
//   - Shows up to 2 decimal places
//   - Trailing zeros after decimal are removed
//   - Whole numbers show no decimal point
//
// Template Usage:
//
//	{{ formatNumber .UserCount }}         <!-- 15234 -> "15,234" -->
//	{{ formatNumber .Revenue }}           <!-- 1234567.89 -> "1,234,567.89" -->
//	{{ .Population | formatNumber }}      <!-- 8419600 -> "8,419,600" -->
//
// Internationalization Note:
//   This function uses US/UK formatting (comma for thousands, period for decimal).
//   For other locales, consider a localized number formatter.
func FormatNumber(n interface{}) string {
	var num float64
	var isFloat bool
	
	// Convert to float64 and track if originally float
	switch v := n.(type) {
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
		isFloat = true
	case float64:
		num = v
		isFloat = true
	default:
		// Not a number, return as string
		return fmt.Sprint(n)
	}
	
	// Handle negative numbers
	sign := ""
	if num < 0 {
		sign = "-"
		num = math.Abs(num)
	}
	
	// Format with appropriate decimal places
	var str string
	if isFloat {
		// For floats, use up to 2 decimal places
		str = fmt.Sprintf("%.2f", num)
		// Remove trailing zeros and decimal point if not needed
		str = strings.TrimRight(str, "0")
		str = strings.TrimRight(str, ".")
	} else {
		// For integers, no decimals
		str = fmt.Sprintf("%.0f", num)
	}
	
	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = "." + parts[1]
	}
	
	// Add commas to integer part
	var result []rune
	intRunes := []rune(intPart)
	for i, r := range intRunes {
		if i > 0 && (len(intRunes)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}
	
	return sign + string(result) + decPart
}