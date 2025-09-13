package testutils

import (
	"reflect"
	"strings"
	"testing"
)

// Check functions - simple, clear names for common assertions

// True fails if condition is false
func True(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		if len(msg) > 0 {
			t.Error(msg[0])
		} else {
			t.Error("Expected true, got false")
		}
	}
}

// False fails if condition is true
func False(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if condition {
		if len(msg) > 0 {
			t.Error(msg[0])
		} else {
			t.Error("Expected false, got true")
		}
	}
}

// Equal fails if values are not equal
func Equal(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("\nExpected: %v\nActual:   %v", expected, actual)
	}
}

// NotEqual fails if values are equal
func NotEqual(t *testing.T, v1, v2 interface{}) {
	t.Helper()
	if reflect.DeepEqual(v1, v2) {
		t.Errorf("Values should not be equal: %v", v1)
	}
}

// Nil fails if value is not nil
func Nil(t *testing.T, value interface{}) {
	t.Helper()
	if !isNil(value) {
		t.Errorf("Expected nil, got %v", value)
	}
}

// NotNil fails if value is nil
func NotNil(t *testing.T, value interface{}) {
	t.Helper()
	if isNil(value) {
		t.Error("Expected non-nil value")
	}
}

// Contains fails if substring is not in string
func Contains(t *testing.T, str, substr string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("Expected string to contain '%s'", substr)
	}
}

// NoError fails if err is not nil
func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Error fails if err is nil
func Error(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("Expected an error but got nil")
	}
}

// Len fails if length doesn't match
func Len(t *testing.T, collection interface{}, expected int) {
	t.Helper()
	v := reflect.ValueOf(collection)
	if v.Len() != expected {
		t.Errorf("Expected length %d, got %d", expected, v.Len())
	}
}