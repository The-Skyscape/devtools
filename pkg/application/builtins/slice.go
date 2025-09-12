package builtins

// Slice extracts a substring from a string using start and end indices.
//
// The function safely handles out-of-bounds indices by clamping them
// to valid ranges.
//
// Parameters:
//   - s: String to slice
//   - start: Starting index (inclusive), clamped to [0, len(s)]
//   - end: Ending index (exclusive), clamped to [start, len(s)]
//
// Returns:
//   Substring from start to end
//
// Examples:
//
//	Slice("Hello World", 0, 5)     // "Hello"
//	Slice("Hello World", 6, 11)    // "World"
//	Slice("Hello World", 6, 100)   // "World" (end clamped to length)
//	Slice("Hello World", -5, 5)    // "Hello" (start clamped to 0)
//	Slice("Hello World", 8, 6)     // "" (start > end returns empty)
//	Slice("Hello", 2, 4)           // "ll"
//	Slice("", 0, 5)                // ""
//
// Index Handling:
//   - Negative start becomes 0
//   - End beyond length becomes length
//   - Start > end returns empty string
//   - Both indices are byte positions, not rune positions
//
// Template Usage:
//
//	{{ slice .Text 0 100 }}         <!-- First 100 characters -->
//	{{ slice .ID 0 8 }}             <!-- First 8 chars of ID -->
//	{{ slice .Path 1 -1 }}          <!-- Won't work for removing first/last -->
//	
// Common Patterns:
//   - Prefix: slice s 0 n
//   - Suffix: slice s n len(s) (but len not available in template)
//   - Middle: slice s start end
//
// Unicode Warning:
//   This function operates on bytes, not runes. For strings with
//   multi-byte Unicode characters, slicing may split characters.
//   Consider a rune-aware slice function for international text.
func Slice(s string, start, end int) string {
	// Clamp start to valid range
	if start < 0 {
		start = 0
	}
	if start > len(s) {
		start = len(s)
	}
	
	// Clamp end to valid range
	if end > len(s) {
		end = len(s)
	}
	if end < start {
		return ""
	}
	
	return s[start:end]
}