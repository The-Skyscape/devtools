package builtins

import "fmt"

// Pluralize returns a properly pluralized string based on count.
//
// This function formats a count with either singular or plural form of a word.
// If no plural form is provided, it adds 's' to the singular form.
//
// Parameters:
//   - count: The number to check for pluralization
//   - singular: The singular form of the word
//   - plural: The plural form (optional, defaults to singular + "s")
//
// Returns:
//   A formatted string with count and appropriate word form
//
// Examples:
//
//	Pluralize(0, "item", "")        // "0 items"
//	Pluralize(1, "item", "")        // "1 item"
//	Pluralize(5, "item", "")        // "5 items"
//	Pluralize(1, "person", "people") // "1 person"
//	Pluralize(3, "person", "people") // "3 people"
//	Pluralize(1, "box", "boxes")    // "1 box"
//	Pluralize(2, "box", "boxes")    // "2 boxes"
//
// Auto-Pluralization:
//   When plural is empty, adds 's' to singular:
//   - "cat" -> "cats"
//   - "dog" -> "dogs"
//
// For irregular plurals, provide explicit plural form:
//   - ("child", "children")
//   - ("mouse", "mice")
//   - ("datum", "data")
//
// Template Usage:
//
//	{{ pluralize .Count "item" "" }}
//	{{ pluralize .Users "user" "" }}
//	{{ pluralize .Geese "goose" "geese" }}
//
// Zero Handling:
//   Zero is treated as plural in English:
//   - 0 items (not "0 item")
//   - 0 people (not "0 person")
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	
	// If no plural form provided, add 's' to singular
	if plural == "" {
		plural = singular + "s"
	}
	
	return fmt.Sprintf("%d %s", count, plural)
}