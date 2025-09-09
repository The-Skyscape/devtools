package application

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"
)

// GetBuiltinFuncs returns the complete template function map with all built-in functions
func GetBuiltinFuncs() template.FuncMap {
	return template.FuncMap{
		// Formatting functions that frontend developers expect
		"formatBytes":     FormatBytes,
		"formatPrice":     FormatPrice,
		"formatPercent":   FormatPercent,
		"formatDuration":  FormatDuration,
		"formatDate":      FormatDate,
		"formatDateTime":  FormatDateTime,
		"formatNumber":    FormatNumber,
		"timeAgo":         TimeAgo,
		
		// String manipulation functions
		"pluralize":  Pluralize,
		"truncate":   Truncate,
		"join":       Join,
		"split":      Split,
		"contains":   Contains,
		"hasPrefix":  HasPrefix,
		"hasSuffix":  HasSuffix,
		"replace":    Replace,
		"title":      strings.Title,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"trim":       strings.TrimSpace,
		
		// Basic math functions (commonly needed)
		"add":   Add,
		"sub":   Subtract,
		"mul":   Multiply,
		"div":   Divide,
		"addf":  AddFloat,
		"subf":  SubtractFloat,
		"mulf":  MultiplyFloat,
		"divf":  DivideFloat,
		
		// Utility functions (without dict/set anti-patterns)
		"toString": ToString,
		"default":  Default,
		"slice":    SliceString,
	}
}

// Math functions

// Add adds two integers or floats
func Add(a, b interface{}) interface{} {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok {
			return va + vb
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va + vb
		}
	}
	return 0
}

// Subtract subtracts two integers or floats
func Subtract(a, b interface{}) interface{} {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok {
			return va - vb
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va - vb
		}
	}
	return 0
}

// Multiply multiplies two numbers
func Multiply(a, b interface{}) interface{} {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok {
			return va * vb
		}
		if vb, ok := b.(float64); ok {
			return float64(va) * vb
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va * vb
		}
		if vb, ok := b.(int); ok {
			return va * float64(vb)
		}
	}
	return 0
}


// Divide divides two numbers
func Divide(a, b interface{}) interface{} {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok && vb != 0 {
			return va / vb
		}
		if vb, ok := b.(float64); ok && vb != 0 {
			return float64(va) / vb
		}
	case float64:
		if vb, ok := b.(float64); ok && vb != 0 {
			return va / vb
		}
		if vb, ok := b.(int); ok && vb != 0 {
			return va / float64(vb)
		}
	}
	return 0
}

// AddFloat adds two numbers as floats
func AddFloat(a, b interface{}) float64 {
	return toFloat(a) + toFloat(b)
}

// SubtractFloat subtracts two numbers as floats
func SubtractFloat(a, b interface{}) float64 {
	return toFloat(a) - toFloat(b)
}

// MultiplyFloat multiplies two numbers as floats
func MultiplyFloat(a, b interface{}) float64 {
	return toFloat(a) * toFloat(b)
}

// DivideFloat divides two numbers as floats
func DivideFloat(a, b interface{}) float64 {
	fb := toFloat(b)
	if fb == 0 {
		return 0
	}
	return toFloat(a) / fb
}

// toFloat is a helper to convert interface{} to float64
func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	}
	return 0
}

// Formatting functions

// FormatBytes formats bytes into human-readable sizes
func FormatBytes(bytes int64) string {
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
func FormatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

// FormatPercent formats a number as a percentage
func FormatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// FormatDuration formats a duration into human-readable format
func FormatDuration(d time.Duration) string {
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
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006")
}

// FormatDateTime formats a time as a date and time string
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}

// FormatNumber formats a number with thousands separators
func FormatNumber(n interface{}) string {
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

// TimeAgo formats a time as "X ago" relative to now
func TimeAgo(t time.Time) string {
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

// String functions

// Pluralize returns singular or plural form based on count
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	if plural == "" {
		plural = singular + "s"
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// Truncate truncates a string to a maximum length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Join joins strings with a separator
func Join(items []string, sep string) string {
	return strings.Join(items, sep)
}

// Split splits a string by separator
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Contains checks if a string contains a substring
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// HasPrefix checks if a string has a prefix
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix checks if a string has a suffix
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Replace replaces all occurrences of a substring
func Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// SliceString returns a substring from start to end
func SliceString(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start > end {
		return ""
	}
	return s[start:end]
}

// Utility functions

// ToString converts any value to a string
func ToString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// Default returns the default value if val is nil, empty, or zero
func Default(def, val interface{}) interface{} {
	if val == nil || val == "" || val == 0 {
		return def
	}
	return val
}

