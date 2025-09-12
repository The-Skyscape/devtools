package builtins

import (
	"fmt"
	"time"
)

// TimeAgo formats a time as a relative string like "5 minutes ago" or "2 days ago".
//
// The function provides human-friendly relative time strings, automatically
// selecting the most appropriate unit based on the time difference.
//
// Parameters:
//   - t: The time to format relative to now
//
// Returns:
//   A relative time string, or "never" for zero time
//
// Time Ranges and Outputs:
//   - Future times: Treated as "just now"
//   - < 1 minute: "just now"
//   - < 1 hour: "X minute(s) ago"
//   - < 24 hours: "X hour(s) ago"  
//   - < 30 days: "X day(s) ago"
//   - < 365 days: "X month(s) ago"
//   - >= 365 days: "X year(s) ago"
//
// Examples:
//
//	now := time.Now()
//	TimeAgo(now)                                   // "just now"
//	TimeAgo(now.Add(-30 * time.Second))           // "just now"
//	TimeAgo(now.Add(-5 * time.Minute))            // "5 minutes ago"
//	TimeAgo(now.Add(-1 * time.Hour))              // "1 hour ago"
//	TimeAgo(now.Add(-25 * time.Hour))             // "1 day ago"
//	TimeAgo(now.Add(-8 * 24 * time.Hour))         // "8 days ago"
//	TimeAgo(now.Add(-45 * 24 * time.Hour))        // "1 month ago"
//	TimeAgo(now.Add(-400 * 24 * time.Hour))       // "1 year ago"
//	TimeAgo(time.Time{})                          // "never"
//
// Singular vs Plural:
//   - Correctly uses singular form for 1 unit ("1 hour ago" not "1 hours ago")
//   - Uses plural for all other values
//
// Precision:
//   - Shows only the largest applicable unit
//   - Does not show fractional units (e.g., "1 hour ago" not "1.5 hours ago")
//   - Months are approximated as 30 days
//   - Years are approximated as 365 days
//
// Template Usage:
//
//	{{ timeAgo .LastSeen }}                <!-- "5 minutes ago" -->
//	{{ .UpdatedAt | timeAgo }}              <!-- "2 days ago" -->
//	Last active: {{ timeAgo .LastActivity }}
//
// Future Times:
//   Future times are displayed as "just now" to handle minor clock skew.
//   For displaying future times, consider a separate function like timeUntil.
func TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	
	duration := time.Since(t)
	
	// Handle future times (clock skew)
	if duration < 0 {
		return "just now"
	}
	
	// Less than a minute
	if duration < time.Minute {
		return "just now"
	}
	
	// Less than an hour - show minutes
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	
	// Less than a day - show hours
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	
	// Less than a month (approximated as 30 days) - show days
	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	
	// Less than a year - show months
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	
	// Years or more
	years := int(duration.Hours() / 24 / 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}