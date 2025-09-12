package builtins

import (
	"fmt"
	"time"
)

// FormatDuration converts a time.Duration into a human-readable string.
//
// The function intelligently formats durations using the most appropriate units,
// showing at most two unit components for clarity (e.g., "2h 30m" not "2h 30m 15s").
//
// Parameters:
//   - d: The duration to format
//
// Returns:
//   A human-readable duration string
//
// Format Rules:
//   - < 1 minute: Shows seconds (e.g., "45s")
//   - < 1 hour: Shows minutes and optional seconds (e.g., "5m 30s")
//   - < 24 hours: Shows hours and optional minutes (e.g., "2h 15m")
//   - >= 24 hours: Shows days and optional hours (e.g., "3d 6h")
//
// Examples:
//
//	FormatDuration(30 * time.Second)              // "30s"
//	FormatDuration(90 * time.Second)              // "1m 30s"
//	FormatDuration(5 * time.Minute)               // "5m"
//	FormatDuration(2*time.Hour + 30*time.Minute)  // "2h 30m"
//	FormatDuration(25 * time.Hour)                // "1d 1h"
//	FormatDuration(72 * time.Hour)                // "3d"
//	FormatDuration(0)                              // "0s"
//	FormatDuration(-30 * time.Second)             // "30s" (absolute value)
//
// Special Cases:
//   - Zero duration returns "0s"
//   - Negative durations are converted to positive
//   - Sub-second durations are shown as "0s"
//   - Trailing zero components are omitted (e.g., "5m" not "5m 0s")
//
// Template Usage:
//
//	{{ formatDuration .Uptime }}
//	{{ formatDuration .ResponseTime }}
//	{{ .BuildDuration | formatDuration }}
func FormatDuration(d time.Duration) string {
	// Handle negative durations
	if d < 0 {
		d = -d
	}
	
	// Handle zero and sub-second durations
	if d < time.Second {
		return "0s"
	}
	
	// Less than a minute - show seconds
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	
	// Less than an hour - show minutes and optional seconds
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}
	
	// Less than a day - show hours and optional minutes
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	
	// Days or more - show days and optional hours
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}