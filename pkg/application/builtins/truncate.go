package builtins

// Truncate shortens a string to a maximum length, adding ellipsis if truncated.
//
// The function ensures strings don't exceed a specified length, adding "..."
// to indicate truncation when necessary.
//
// Parameters:
//   - s: The string to truncate
//   - maxLen: Maximum length including ellipsis
//
// Returns:
//   Truncated string with ellipsis if needed
//
// Examples:
//
//	Truncate("Hello World", 5)   // "He..."
//	Truncate("Hello World", 11)  // "Hello World" (no truncation needed)
//	Truncate("Hello World", 8)   // "Hello..."
//	Truncate("Hi", 5)           // "Hi" (no truncation needed)
//	Truncate("Test", 3)         // "Tes" (maxLen <= 3, no ellipsis)
//	Truncate("", 10)            // ""
//
// Ellipsis Handling:
//   - If maxLen <= 3: Truncates without ellipsis
//   - If maxLen > 3: Reserves 3 characters for "..."
//   - No truncation if string fits within maxLen
//
// Unicode Safety:
//   The function counts bytes, not runes. For proper Unicode handling,
//   consider a rune-aware truncation function.
//
// Template Usage:
//
//	{{ truncate .Description 100 }}
//	{{ .Title | truncate 50 }}
//	{{ truncate .Content 200 }}      <!-- preview text -->
//
// Common Use Cases:
//   - Preview text for articles
//   - Limiting display in tables
//   - Creating text summaries
//   - Preventing layout breaks
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	// For very short max lengths, just truncate without ellipsis
	if maxLen <= 3 {
		return s[:maxLen]
	}
	
	// Truncate and add ellipsis
	return s[:maxLen-3] + "..."
}