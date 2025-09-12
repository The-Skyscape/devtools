package builtins

import "fmt"

// FormatBytes converts a byte count into a human-readable string with
// appropriate unit suffixes (B, KB, MB, GB, TB, PB, EB).
//
// The function uses binary units (1024 bytes = 1 KB) rather than decimal
// units (1000 bytes = 1 KB) following common computing conventions.
//
// Parameters:
//   - bytes: The number of bytes to format. Negative values are treated as positive.
//
// Returns:
//   A formatted string with one decimal place precision for values >= 1 KB.
//
// Examples:
//
//	FormatBytes(0)          // "0 B"
//	FormatBytes(1023)       // "1023 B"
//	FormatBytes(1024)       // "1.0 KB"
//	FormatBytes(1536)       // "1.5 KB"
//	FormatBytes(1048576)    // "1.0 MB"
//	FormatBytes(5368709120) // "5.0 GB"
//
// Edge Cases:
//   - Negative values are converted to positive
//   - Values less than 1024 show as bytes with no decimal
//   - Maximum unit shown is EB (exabytes)
//
// Template Usage:
//
//	{{ formatBytes .FileSize }}
//	{{ formatBytes 1048576 }} <!-- outputs: 1.0 MB -->
func FormatBytes(bytes int64) string {
	const unit = 1024
	
	// Handle negative values
	if bytes < 0 {
		bytes = -bytes
	}
	
	// Less than 1 KB - show as bytes
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	// Calculate the appropriate unit
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
		// Prevent overflow for extremely large values
		if exp >= 6 { // Stop at EB (exabytes)
			break
		}
	}
	
	// Format with appropriate unit suffix
	unitSuffix := "KMGTPE"[exp] // Kilo, Mega, Giga, Tera, Peta, Exa
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), unitSuffix)
}