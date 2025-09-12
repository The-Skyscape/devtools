package builtins

import "strings"

// HasPrefix checks if a string starts with a given prefix.
//
// This is a direct wrapper around strings.HasPrefix for template use.
// The check is case-sensitive.
//
// Parameters:
//   - s: String to check
//   - prefix: Prefix to look for
//
// Returns:
//   true if s starts with prefix, false otherwise
//
// Examples:
//
//	HasPrefix("Hello World", "Hello")   // true
//	HasPrefix("Hello World", "hello")   // false (case-sensitive)
//	HasPrefix("Hello World", "World")   // false
//	HasPrefix("Hello World", "")        // true (empty prefix always matches)
//	HasPrefix("", "test")               // false
//	HasPrefix("", "")                   // true
//	HasPrefix("/api/users", "/api")     // true
//
// Template Usage:
//
//	{{ if hasPrefix .URL "https://" }}
//	  <span class="secure">Secure</span>
//	{{ end }}
//	
//	{{ if hasPrefix .Path "/admin" }}
//	  <!-- Admin section -->
//	{{ end }}
//	
//	{{ if hasPrefix .Email "noreply@" }}
//	  <em>Do not reply to this email</em>
//	{{ end }}
//
// Common Patterns:
//   - URL scheme check: hasPrefix url "https://"
//   - Path routing: hasPrefix path "/api/"
//   - File type check: hasPrefix filename "temp_"
//   - Command detection: hasPrefix input "/"
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}