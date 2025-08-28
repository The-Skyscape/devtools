package application

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"
)

// FormatHelpers provides template formatting functions
type FormatHelpers struct{}

// NewFormatHelpers creates a new format helper instance
func NewFormatHelpers() *FormatHelpers {
	return &FormatHelpers{}
}

// FuncMap returns the template function map with all formatting helpers
func (f *FormatHelpers) FuncMap() template.FuncMap {
	return template.FuncMap{
		"formatBytes":     f.FormatBytes,
		"formatPrice":     f.FormatPrice,
		"formatPercent":   f.FormatPercent,
		"formatDuration":  f.FormatDuration,
		"formatDate":      f.FormatDate,
		"formatDateTime":  f.FormatDateTime,
		"formatNumber":    f.FormatNumber,
		"pluralize":       f.Pluralize,
		"truncate":        f.Truncate,
		"join":            f.Join,
		"split":           f.Split,
		"contains":        f.Contains,
		"hasPrefix":       f.HasPrefix,
		"hasSuffix":       f.HasSuffix,
		"replace":         f.Replace,
		"title":           strings.Title,
		"lower":           strings.ToLower,
		"upper":           strings.ToUpper,
		"trim":            strings.TrimSpace,
	}
}

// FormatBytes formats bytes into human-readable sizes
func (f *FormatHelpers) FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatPrice formats a number as currency
func (f *FormatHelpers) FormatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

// FormatPercent formats a number as a percentage
func (f *FormatHelpers) FormatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// FormatDuration formats a duration into human-readable format
func (f *FormatHelpers) FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}

// FormatDate formats a time as a date string
func (f *FormatHelpers) FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006")
}

// FormatDateTime formats a time as a date and time string
func (f *FormatHelpers) FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}

// FormatNumber formats a number with thousands separators
func (f *FormatHelpers) FormatNumber(n interface{}) string {
	var num float64
	switch v := n.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	default:
		return fmt.Sprint(n)
	}
	
	// Handle negative numbers
	sign := ""
	if num < 0 {
		sign = "-"
		num = math.Abs(num)
	}
	
	// Format with commas
	parts := strings.Split(fmt.Sprintf("%.2f", num), ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 && parts[1] != "00" {
		decPart = "." + parts[1]
	}
	
	// Add commas to integer part
	var result []rune
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}
	
	return sign + string(result) + decPart
}

// Pluralize returns singular or plural form based on count
func (f *FormatHelpers) Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	if plural == "" {
		plural = singular + "s"
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// Truncate truncates a string to a maximum length
func (f *FormatHelpers) Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Join joins strings with a separator
func (f *FormatHelpers) Join(items []string, sep string) string {
	return strings.Join(items, sep)
}

// Split splits a string by separator
func (f *FormatHelpers) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Contains checks if a string contains a substring
func (f *FormatHelpers) Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// HasPrefix checks if a string has a prefix
func (f *FormatHelpers) HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix checks if a string has a suffix
func (f *FormatHelpers) HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Replace replaces all occurrences of a substring
func (f *FormatHelpers) Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// TimeAgo formats a time as "X ago" relative to now
func (f *FormatHelpers) TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	
	years := int(duration.Hours() / 24 / 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// RegisterHelpers registers all formatting helpers with the template engine
func RegisterHelpers(tmpl *template.Template) *template.Template {
	helpers := NewFormatHelpers()
	return tmpl.Funcs(helpers.FuncMap())
}