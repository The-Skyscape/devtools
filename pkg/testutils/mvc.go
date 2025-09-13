package testutils

import (
	"net/http/httptest"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// ControllerTest provides context for testing a specific controller
type ControllerTest[H application.Handler] struct {
	*testing.T
	prefix     string
	controller H
}

// TestController tests a controller and runs additional tests with the controller instance
func TestController[H application.Handler](t *testing.T, prefix string, controller H, additionalTests ...func(*ControllerTest[H])) {
	t.Run("Controller_"+prefix, func(t *testing.T) {
		// Create controller test context
		ct := &ControllerTest[H]{
			T:          t,
			prefix:     prefix,
			controller: controller,
		}

		t.Run("RequestIsolation", func(t *testing.T) {
			req1 := httptest.NewRequest("GET", "/test/1", nil)
			req2 := httptest.NewRequest("GET", "/test/2", nil)

			handler1 := controller.Handle(req1)
			handler2 := controller.Handle(req2)

			if handler1 == nil || handler2 == nil {
				t.Fatal("Handle should return handler instances")
			}

			// Verify they are different instances (value receiver pattern)
			if handler1 == handler2 {
				t.Error("Handle should return different instances for request isolation")
			}
		})

		// Run additional tests with controller instance
		for _, test := range additionalTests {
			test(ct)
		}
	})
}

// TestMethod verifies a method exists and runs a test for it
func (ct *ControllerTest[H]) TestMethod(methodName string, f func(*testing.T, H)) {
	ct.T.Run(methodName, func(t *testing.T) {
		// Check method exists for template access
		AssertHasMethod(t, ct.controller, methodName)
		
		// Run the test with the typed controller
		f(t, ct.controller)
	})
}

// Run runs a subtest with the typed controller instance available
func (ct *ControllerTest[H]) Run(name string, f func(*testing.T, H)) {
	ct.T.Run(name, func(t *testing.T) {
		f(t, ct.controller)
	})
}

// Prefix returns the controller's prefix
func (ct *ControllerTest[H]) Prefix() string {
	return ct.prefix
}
