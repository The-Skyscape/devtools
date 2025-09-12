// Package builtins provides a comprehensive set of type-safe template functions
// for use in Go HTML templates. Inspired by lodash's documentation style and
// commitment to utility, this package offers thoroughly documented, well-tested
// functions for common template operations.
//
// The package is organized into four main categories:
//
// Format Functions:
// Functions for formatting data into human-readable strings, including bytes,
// currency, percentages, dates, and relative time.
//
// Math Functions:
// Type-safe arithmetic operations with proper handling of integers and floats,
// including division by zero protection.
//
// String Functions:
// Common string manipulation utilities like truncation, pluralization, and
// case conversion with Unicode support.
//
// Utility Functions:
// General-purpose helpers for type conversion and default values.
//
// Type Safety:
// All functions are designed with type safety in mind. Unlike generic any
// approaches, functions explicitly handle expected types and return predictable
// zero values for invalid inputs rather than panicking.
//
// Template Usage:
// These functions are designed to be registered with html/template.FuncMap:
//
//	funcs := builtins.FuncMap
//	tmpl := template.New("").Funcs(funcs).Parse(templateString)
//
// Error Handling:
// Functions follow Go's principle of graceful degradation. Invalid inputs
// return sensible zero values rather than errors, making them safe for use
// in templates where error handling is limited.
package builtins

import "html/template"

// FuncMap contains all builtin functions as a template.FuncMap
// ready to be used with html/template or text/template.
//
// Example:
//
//	tmpl := template.New("example").Funcs(builtins.FuncMap)
//	tmpl.Parse(`Price: {{ formatPrice 19.99 }}`)
var FuncMap = template.FuncMap{
		// Format functions
		"formatBytes":    FormatBytes,
		"formatPrice":    FormatPrice,
		"formatPercent":  FormatPercent,
		"formatDuration": FormatDuration,
		"formatDate":     FormatDate,
		"formatDateTime": FormatDateTime,
		"formatNumber":   FormatNumber,
		"timeAgo":        TimeAgo,

		// String functions
		"pluralize": Pluralize,
		"truncate":  Truncate,
		"join":      Join,
		"split":     Split,
		"contains":  Contains,
		"hasPrefix":  HasPrefix,
		"hasSuffix":  HasSuffix,
		"replace":    Replace,
		"replaceAll": ReplaceAll,
		"slice":      Slice,
		"lower":     Lower,
		"upper":     Upper,
		"title":     Title,
		"trim":      Trim,

		// Math functions
		"add":   Add,
		"sub":   Subtract,
		"mul":   Multiply,
		"div":   Divide,
		"addf":  AddFloat,
		"subf":  SubtractFloat,
		"mulf":  MultiplyFloat,
		"divf":  DivideFloat,
		"mod":   Mod,
		"max":   Max,
		"min":   Min,
		"round": Round,
		"floor": Floor,
		"ceil":  Ceil,

		// Utility functions
		"toString": ToString,
		"default":  Default,
		"coalesce": Coalesce,
		"isEmpty":  IsEmpty,
		"isNil":    IsNil,
}
