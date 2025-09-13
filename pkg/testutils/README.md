# TestUtils Package

Testing utilities for DevTools applications, providing helpers for testing controllers, database operations, HTTP requests, and assertions.

## Controller Testing

### Quick Start

```go
import (
    "testing"
    "github.com/The-Skyscape/devtools/pkg/testutils"
)

func TestMyController(t *testing.T) {
    // Setup test database
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(t, db)
    
    // Create your application
    app := application.New()
    
    // Define test cases
    tests := []testutils.ControllerTest{
        {
            Name:           "GET home page",
            Method:         "GET",
            Path:           "/",
            ExpectedStatus: http.StatusOK,
            ExpectedBody:   "Welcome",
        },
    }
    
    // Run tests
    testutils.TestController(t, app, tests)
}
```

### Testing Patterns

#### 1. Table-Driven Tests with TestController

Best for testing multiple scenarios:

```go
tests := []testutils.ControllerTest{
    {
        Name:           "Valid form submission",
        Method:         "POST",
        Path:           "/submit",
        Body:           url.Values{"name": []string{"John"}},
        ExpectedStatus: http.StatusOK,
        ExpectedBody:   "Success",
    },
    {
        Name:           "Invalid form - missing name",
        Method:         "POST",
        Path:           "/submit",
        Body:           url.Values{},
        ExpectedStatus: http.StatusBadRequest,
        ExpectedBody:   "Name required",
    },
}

testutils.TestController(t, app, tests)
```

#### 2. Fluent API with ControllerTestSuite

Best for readable, chainable assertions:

```go
suite := testutils.NewControllerTestSuite(t, app)

suite.
    GET("/about").
    AssertOK().
    AssertContains("About Us").
    AssertNotContains("Error")

suite.
    POST("/login", url.Values{
        "username": []string{"admin"},
        "password": []string{"secret"},
    }).
    AssertRedirect("/dashboard")
```

#### 3. Direct Handler Testing

Best for unit testing individual handlers:

```go
handler := func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello"))
}

testutils.TestControllerMethod(t, handler, "GET", "/", nil,
    func(t *testing.T, w *testutils.MockResponse) {
        w.AssertOK(t)
        w.AssertBodyContains(t, "Hello")
    })
```

### Testing with Authentication

```go
// Create authenticated request
token := "your-jwt-token"
req := testutils.MockAuthenticatedRequest("GET", "/profile", nil, token)

// Or with the suite
suite.
    GET("/profile").
    WithHeader("Authorization", "Bearer "+token).
    AssertOK()
```

### Testing HTMX Endpoints

```go
// Test HTMX refresh response
handler := func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("HX-Refresh", "true")
    w.WriteHeader(http.StatusOK)
}

req := testutils.MockRequest("POST", "/update", nil)
req.Header.Set("HX-Request", "true")

w := httptest.NewRecorder()
handler(w, req)

testutils.AssertHTMXRefresh(t, w)
```

### Custom Assertions

Use the `CheckFunc` field for custom assertions:

```go
{
    Name:   "Custom validation",
    Method: "GET",
    Path:   "/api/data",
    CheckFunc: func(t *testing.T, w *testutils.MockResponse) {
        var data map[string]any
        w.AssertJSON(t, &data)
        
        // Custom assertions
        if data["count"].(float64) < 10 {
            t.Error("Expected at least 10 items")
        }
    },
}
```

### Setup and Cleanup

Add setup and cleanup functions to test cases:

```go
{
    Name:   "Test with database setup",
    Method: "GET",
    Path:   "/users",
    SetupFunc: func() {
        // Create test users
        testutils.CreateTestUser(t, db, "user1@example.com")
        testutils.CreateTestUser(t, db, "user2@example.com")
    },
    CleanupFunc: func() {
        // Cleanup happens automatically with test DB
    },
    ExpectedBody: "2 users",
}
```

## HTTP Testing Utilities

### MockRequest Functions

```go
// Basic request
req := testutils.MockRequest("GET", "/path", nil)

// With JSON body
req := testutils.MockJSONRequest("POST", "/api", struct{Name string}{Name: "test"})

// With form data
req := testutils.MockFormRequest("POST", "/form", url.Values{"key": []string{"value"}})

// With authentication
req := testutils.MockAuthenticatedRequest("GET", "/secure", nil, "token")
```

### MockResponse Assertions

```go
w := testutils.NewMockResponse()

// Status assertions
w.AssertOK(t)                          // 200 OK
w.AssertStatus(t, http.StatusCreated)  // Any status

// Body assertions
w.AssertBodyContains(t, "expected text")
w.AssertJSON(t, &result)

// Header assertions
w.AssertHeader(t, "Content-Type", "application/json")

// Redirect assertions
w.AssertRedirect(t, "/new-location")
```

## Database Testing

### Test Database Setup

```go
func TestWithDatabase(t *testing.T) {
    // Setup temporary test database
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(t, db)
    
    // Database is ready to use
    // Tables are automatically created
}
```

### Test Data Creation

```go
// Create test user
user := testutils.CreateTestUser(t, db, "test@example.com")

// Create custom test data
testutils.CreateTestData(t, db, "users", map[string]any{
    "name":  "John",
    "email": "john@example.com",
})
```

## Assertions

Basic assertions for all test types:

```go
testutils.AssertEqual(t, expected, actual)
testutils.AssertNotEqual(t, unexpected, actual)
testutils.AssertNil(t, value)
testutils.AssertNotNil(t, value)
testutils.AssertTrue(t, condition)
testutils.AssertFalse(t, condition)
testutils.AssertError(t, err)
testutils.AssertNoError(t, err)
testutils.AssertContains(t, haystack, needle)
testutils.AssertNotContains(t, haystack, needle)
```

## Best Practices

1. **Use Table-Driven Tests**: Group related test cases together
2. **Test Happy Path and Errors**: Always test both success and failure cases
3. **Use Descriptive Names**: Make test names describe what they're testing
4. **Cleanup Resources**: Always defer cleanup functions
5. **Mock External Dependencies**: Use mock HTTP clients for external services
6. **Test One Thing**: Each test should verify a single behavior
7. **Use Helper Functions**: Extract common setup into helper functions

## Example: Complete Controller Test

```go
func TestUserController(t *testing.T) {
    // Setup
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(t, db)
    
    app := application.New()
    
    // Create test data
    user := testutils.CreateTestUser(t, db, "test@example.com")
    token := generateTestToken(user)
    
    // Define comprehensive test cases
    tests := []testutils.ControllerTest{
        {
            Name:           "List users - authenticated",
            Method:         "GET",
            Path:           "/users",
            Headers:        map[string]string{"Authorization": "Bearer " + token},
            ExpectedStatus: http.StatusOK,
            CheckFunc: func(t *testing.T, w *testutils.MockResponse) {
                var users []map[string]any
                w.AssertJSON(t, &users)
                testutils.AssertEqual(t, 1, len(users))
            },
        },
        {
            Name:           "List users - unauthenticated",
            Method:         "GET",
            Path:           "/users",
            ExpectedStatus: http.StatusUnauthorized,
        },
        {
            Name:           "Create user - valid",
            Method:         "POST",
            Path:           "/users",
            Headers:        map[string]string{"Authorization": "Bearer " + token},
            Body:           map[string]string{"email": "new@example.com"},
            ExpectedStatus: http.StatusCreated,
            CheckFunc: func(t *testing.T, w *testutils.MockResponse) {
                location := w.Header().Get("Location")
                testutils.AssertContains(t, location, "/users/")
            },
        },
        {
            Name:           "Create user - invalid email",
            Method:         "POST",
            Path:           "/users",
            Headers:        map[string]string{"Authorization": "Bearer " + token},
            Body:           map[string]string{"email": "invalid"},
            ExpectedStatus: http.StatusBadRequest,
            ExpectedBody:   "Invalid email",
        },
    }
    
    testutils.TestController(t, app, tests)
}
```

This testing infrastructure provides everything needed to thoroughly test DevTools applications!