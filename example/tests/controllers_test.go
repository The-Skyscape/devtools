package tests

import (
	"testing"

	"github.com/The-Skyscape/devtools/example/controllers"
	"github.com/The-Skyscape/devtools/pkg/testutils"
)

func TestDucksController(t *testing.T) {
	prefix, controller := controllers.Ducks()
	
	// Inject test database collection
	controller.DucksCollection = TestDucks

	testutils.TestController(t, prefix, controller, func(ct *testutils.ControllerTest[*controllers.DucksController]) {
		ct.TestMethod("AllDucks", func(t *testing.T, c *controllers.DucksController) {
			allDucks := c.AllDucks()

			// Should have 3 test ducks from setup
			if len(allDucks) != 3 {
				t.Errorf("Expected 3 ducks, got %d", len(allDucks))
			}
		})

		ct.TestMethod("CountDucks", func(t *testing.T, c *controllers.DucksController) {
			count := c.CountDucks()
			if count != 3 {
				t.Errorf("CountDucks should return 3, got %d", count)
			}
		})

		ct.TestMethod("IsEmpty", func(t *testing.T, c *controllers.DucksController) {
			isEmpty := c.IsEmpty()
			if isEmpty {
				t.Error("IsEmpty should return false when ducks exist")
			}

			// Clear ducks and test again
			ClearData()
			isEmpty = c.IsEmpty()
			if !isEmpty {
				t.Error("IsEmpty should return true when no ducks exist")
			}

			// Restore test data for other tests
			SetupTestData()
		})
	})
}
