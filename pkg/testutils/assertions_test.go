package testutils

import (
	"errors"
	"testing"
	"time"
)

func TestAssertEqual(t *testing.T) {
	// Test with equal values
	AssertEqual(t, 42, 42)
	AssertEqual(t, "hello", "hello")
	AssertEqual(t, []int{1, 2, 3}, []int{1, 2, 3})

	// Test with custom message
	AssertEqual(t, true, true, "Booleans should be equal")
}

func TestAssertNotEqual(t *testing.T) {
	// Test with different values
	AssertNotEqual(t, 42, 43)
	AssertNotEqual(t, "hello", "world")
	AssertNotEqual(t, []int{1, 2, 3}, []int{3, 2, 1})
}

func TestAssertNil(t *testing.T) {
	// Test with nil values
	var ptr *string
	AssertNil(t, ptr)
	AssertNil(t, nil)

	var err error
	AssertNil(t, err)

	var slice []int
	AssertNil(t, slice)

	var m map[string]int
	AssertNil(t, m)
}

func TestAssertNotNil(t *testing.T) {
	// Test with non-nil values
	str := "test"
	AssertNotNil(t, &str)
	AssertNotNil(t, errors.New("test error"))
	AssertNotNil(t, []int{1, 2, 3})
	AssertNotNil(t, make(map[string]int))
}

func TestAssertTrue(t *testing.T) {
	AssertTrue(t, true)
	AssertTrue(t, 1 == 1)
	AssertTrue(t, len("hello") == 5)
}

func TestAssertFalse(t *testing.T) {
	AssertFalse(t, false)
	AssertFalse(t, 1 == 2)
	AssertFalse(t, len("hello") == 4)
}

func TestAssertError(t *testing.T) {
	err := errors.New("test error")
	AssertError(t, err)

	// Test with custom error types
	AssertError(t, errors.New("another error"))
}

func TestAssertNoError(t *testing.T) {
	var err error
	AssertNoError(t, err)

	// Test with function that returns no error
	err = func() error { return nil }()
	AssertNoError(t, err)
}

func TestAssertContains(t *testing.T) {
	AssertContains(t, "hello world", "world")
	AssertContains(t, "testing is fun", "test")
	AssertContains(t, "abcdefg", "cde")

	// Test edge cases
	AssertContains(t, "test", "test")
	AssertContains(t, "a", "a")
}

func TestAssertTimeBetween(t *testing.T) {
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	AssertTimeBetween(t, now, start, end)

	// Test boundary values
	AssertTimeBetween(t, start, start, end)
	AssertTimeBetween(t, end, start, end)
}

func TestAssertTimeWithin(t *testing.T) {
	now := time.Now()
	expected := now.Add(100 * time.Millisecond)
	tolerance := 200 * time.Millisecond

	AssertTimeWithin(t, now, expected, tolerance)

	// Test exact match
	AssertTimeWithin(t, now, now, 0)

	// Test with larger tolerance
	future := now.Add(1 * time.Second)
	AssertTimeWithin(t, future, now, 2*time.Second)
}

func TestAssertLen(t *testing.T) {
	// Test with slice
	AssertLen(t, []int{1, 2, 3}, 3)
	AssertLen(t, []string{}, 0)

	// Test with array
	arr := [5]int{1, 2, 3, 4, 5}
	AssertLen(t, arr, 5)

	// Test with string
	AssertLen(t, "hello", 5)
	AssertLen(t, "", 0)

	// Test with map
	m := map[string]int{"a": 1, "b": 2}
	AssertLen(t, m, 2)
	AssertLen(t, map[string]int{}, 0)
}

func TestIsNil(t *testing.T) {
	// Test various nil types
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"nil interface", nil, true},
		{"nil pointer", (*string)(nil), true},
		{"nil slice", ([]int)(nil), true},
		{"nil map", (map[string]int)(nil), true},
		{"nil channel", (chan int)(nil), true},
		{"nil func", (func())(nil), true},
		{"non-nil string pointer", new(string), false},
		{"non-nil slice", []int{}, false},
		{"non-nil map", make(map[string]int), false},
		{"int zero value", 0, false},
		{"string zero value", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNil(tt.value)
			if result != tt.expected {
				t.Errorf("isNil(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "World", false}, // Case sensitive
		{"testing", "test", true},
		{"testing", "ing", true},
		{"testing", "testing", true},
		{"testing", "testings", false},
		{"", "", false}, // Empty needle
		{"test", "", false},
		{"", "test", false},
		{"a", "a", true},
		{"ab", "b", true},
		{"abc", "bc", true},
	}

	for _, tt := range tests {
		t.Run(tt.haystack+"/"+tt.needle, func(t *testing.T) {
			result := contains(tt.haystack, tt.needle)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v",
					tt.haystack, tt.needle, result, tt.expected)
			}
		})
	}
}
