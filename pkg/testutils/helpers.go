package testutils

import (
	"reflect"
	"testing"
)

// AssertHasMethod checks if a struct has a method with the given name
// and warns if the method returns an error as second value (which can halt templates)
func AssertHasMethod(t *testing.T, obj interface{}, methodName string) {
	t.Helper()
	
	v := reflect.ValueOf(obj)
	method := v.MethodByName(methodName)
	
	if !method.IsValid() {
		t.Errorf("Expected method '%s' not found on %T", methodName, obj)
		return
	}
	
	// Check if method returns an error as second value
	methodType := method.Type()
	if methodType.NumOut() == 2 {
		// Check if second return type is error
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if methodType.Out(1).Implements(errorInterface) {
			t.Logf("WARNING: Method '%s' returns (value, error). This will halt template execution on error. Consider returning only the value with a zero/default value on error.", methodName)
		}
	}
}

// AssertMethodExists is an alias for AssertHasMethod for backward compatibility
func AssertMethodExists(t *testing.T, obj interface{}, methodName string) {
	AssertHasMethod(t, obj, methodName)
}