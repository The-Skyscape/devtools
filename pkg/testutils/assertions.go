package testutils

import (
	"reflect"
	"testing"
	"time"
)

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message ...string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := "Values not equal"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nExpected: %v\nActual: %v", msg, expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}, message ...string) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		msg := "Values are equal"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nValue: %v", msg, actual)
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, value interface{}, message ...string) {
	t.Helper()
	if !isNil(value) {
		msg := "Expected nil"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nGot: %v", msg, value)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value interface{}, message ...string) {
	t.Helper()
	if isNil(value) {
		msg := "Expected non-nil value"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

// AssertTrue checks if a boolean is true
func AssertTrue(t *testing.T, value bool, message ...string) {
	t.Helper()
	if !value {
		msg := "Expected true"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

// AssertFalse checks if a boolean is false
func AssertFalse(t *testing.T, value bool, message ...string) {
	t.Helper()
	if value {
		msg := "Expected false"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

// AssertError checks if an error occurred
func AssertError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err == nil {
		msg := "Expected error but got nil"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

// AssertNoError checks if no error occurred
func AssertNoError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err != nil {
		msg := "Unexpected error"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s: %v", msg, err)
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, haystack, needle string, message ...string) {
	t.Helper()
	if !contains(haystack, needle) {
		msg := "String does not contain expected substring"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nString: %s\nExpected to contain: %s", msg, haystack, needle)
	}
}

// AssertTimeBetween checks if a time is between two other times
func AssertTimeBetween(t *testing.T, actual, start, end time.Time, message ...string) {
	t.Helper()
	if actual.Before(start) || actual.After(end) {
		msg := "Time not in expected range"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nActual: %v\nExpected between: %v and %v", msg, actual, start, end)
	}
}

// AssertTimeWithin checks if a time is within a duration of another time
func AssertTimeWithin(t *testing.T, actual, expected time.Time, tolerance time.Duration, message ...string) {
	t.Helper()
	diff := actual.Sub(expected)
	if diff < -tolerance || diff > tolerance {
		msg := "Time not within tolerance"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nActual: %v\nExpected: %v ±%v\nDifference: %v", msg, actual, expected, tolerance, diff)
	}
}

// AssertLen checks the length of a slice, array, map, or string
func AssertLen(t *testing.T, value interface{}, expected int, message ...string) {
	t.Helper()
	actual := reflect.ValueOf(value).Len()
	if actual != expected {
		msg := "Length mismatch"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nExpected length: %d\nActual length: %d", msg, expected, actual)
	}
}

// Helper functions

func isNil(value interface{}) bool {
	if value == nil {
		return true
	}
	
	// Check for typed nil values
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && 
		(haystack == needle || len(haystack) > len(needle) && containsSubstring(haystack, needle))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}