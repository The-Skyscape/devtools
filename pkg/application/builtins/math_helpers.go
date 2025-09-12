package builtins

import (
	"fmt"
	"math"
)

// toInt safely converts an any to int if possible.
// Returns the int value and true if successful, 0 and false otherwise.
func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int32:
		return int(val), true
	case int64:
		// Check for overflow
		if val >= math.MinInt && val <= math.MaxInt {
			return int(val), true
		}
		return 0, false
	case float32:
		// Only convert if it's a whole number
		if val == float32(int(val)) {
			return int(val), true
		}
		return 0, false
	case float64:
		// Only convert if it's a whole number
		if val == math.Trunc(val) {
			return int(val), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toFloat64 converts an any to float64.
// Returns 0 if the value cannot be converted.
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		// Try to parse string as number
		var f float64
		if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

// isNumeric checks if a value is a numeric type.
func isNumeric(v any) bool {
	switch v.(type) {
	case int, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}
