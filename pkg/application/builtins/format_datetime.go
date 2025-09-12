package builtins

import "time"

// FormatDateTime formats a time.Time as a human-readable date and time string.
//
// The function uses a clean 12-hour format: "Jan 2, 2006 3:04 PM" which
// displays as "Feb 14, 2024 3:30 PM" for an afternoon time.
//
// Parameters:
//   - t: The time to format
//
// Returns:
//   A formatted date and time string, or empty string for zero time
//
// Examples:
//
//	t1 := time.Date(2024, 2, 14, 15, 30, 0, 0, time.UTC)
//	FormatDateTime(t1)                // "Feb 14, 2024 3:30 PM"
//	
//	t2 := time.Date(2024, 12, 25, 9, 0, 0, 0, time.UTC)
//	FormatDateTime(t2)                // "Dec 25, 2024 9:00 AM"
//	
//	t3 := time.Date(2024, 1, 1, 0, 15, 0, 0, time.UTC)
//	FormatDateTime(t3)                // "Jan 1, 2024 12:15 AM"
//	
//	FormatDateTime(time.Time{})       // "" (empty for zero time)
//	FormatDateTime(time.Now())        // e.g., "Jan 15, 2024 2:45 PM"
//
// Format Details:
//   - Date: "Jan 2, 2006" format
//   - Time: 12-hour format with AM/PM
//   - Minutes: Always shown with two digits
//   - Seconds: Not displayed
//   - Timezone: Not displayed (uses time's location)
//
// Zero Value Handling:
//   - Returns empty string for zero time
//   - Useful for optional timestamps in templates
//
// Template Usage:
//
//	{{ formatDateTime .LastModified }}
//	{{ .CreatedAt | formatDateTime }}
//	{{ if .ScheduledAt }}Scheduled for: {{ formatDateTime .ScheduledAt }}{{ end }}
//
// Time Zone Considerations:
//   The displayed time uses the time.Time's location. To display in a
//   specific timezone, convert the time before passing to this function:
//
//	{{ .CreatedAt.In .UserTimezone | formatDateTime }}
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}