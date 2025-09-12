package builtins

import "time"

// FormatDate formats a time.Time as a human-readable date string.
//
// The function uses a clean, unambiguous format: "Jan 2, 2006" which
// displays as "Feb 14, 2024" for Valentine's Day 2024.
//
// Parameters:
//   - t: The time to format
//
// Returns:
//   A formatted date string, or empty string for zero time
//
// Examples:
//
//	t1 := time.Date(2024, 2, 14, 15, 30, 0, 0, time.UTC)
//	FormatDate(t1)                    // "Feb 14, 2024"
//	
//	t2 := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
//	FormatDate(t2)                    // "Dec 25, 2024"
//	
//	FormatDate(time.Time{})           // "" (empty for zero time)
//	FormatDate(time.Now())            // e.g., "Jan 15, 2024"
//
// Format Details:
//   - Month: Abbreviated month name (Jan, Feb, Mar, etc.)
//   - Day: Day of month without leading zero
//   - Year: 4-digit year
//   - Time and timezone are ignored
//
// Zero Value Handling:
//   - Returns empty string for zero time
//   - Useful for optional dates in templates
//
// Template Usage:
//
//	{{ formatDate .CreatedAt }}
//	{{ .PublishedAt | formatDate }}
//	{{ if .DueDate }}Due: {{ formatDate .DueDate }}{{ end }}
//
// Localization Note:
//   Currently uses English month names. For localization,
//   consider using a localized format function instead.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006")
}