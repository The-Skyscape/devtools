package testutils

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestMockRequest(t *testing.T) {
	req := MockRequest("GET", "/test", nil)
	
	AssertEqual(t, "GET", req.Method)
	AssertEqual(t, "/test", req.URL.Path)
	AssertNil(t, req.Body)
	
	// Test with body
	body := strings.NewReader("test body")
	req = MockRequest("POST", "/api/data", body)
	AssertEqual(t, "POST", req.Method)
	AssertNotNil(t, req.Body)
}

func TestMockJSONRequest(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"value": 42,
	}
	
	req := MockJSONRequest("POST", "/api/json", data)
	
	AssertEqual(t, "POST", req.Method)
	AssertEqual(t, "/api/json", req.URL.Path)
	AssertEqual(t, "application/json", req.Header.Get("Content-Type"))
	AssertNotNil(t, req.Body)
}

func TestMockFormRequest(t *testing.T) {
	values := url.Values{
		"username": []string{"testuser"},
		"password": []string{"secret"},
	}
	
	req := MockFormRequest("POST", "/login", values)
	
	AssertEqual(t, "POST", req.Method)
	AssertEqual(t, "/login", req.URL.Path)
	AssertEqual(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
	AssertNotNil(t, req.Body)
}

func TestMockResponse(t *testing.T) {
	resp := NewMockResponse()
	
	// Test initial state
	AssertEqual(t, 200, resp.Code) // Default is 200
	AssertEqual(t, "", resp.Body.String())
	
	// Write response
	resp.WriteHeader(http.StatusCreated)
	resp.Write([]byte("test response"))
	
	AssertEqual(t, http.StatusCreated, resp.Code)
	AssertEqual(t, "test response", resp.Body.String())
}

func TestMockResponseAssertions(t *testing.T) {
	t.Run("AssertStatus", func(t *testing.T) {
		resp := NewMockResponse()
		resp.WriteHeader(http.StatusNotFound)
		resp.AssertStatus(t, http.StatusNotFound)
	})
	
	t.Run("AssertOK", func(t *testing.T) {
		resp := NewMockResponse()
		resp.WriteHeader(http.StatusOK)
		resp.AssertOK(t)
	})
	
	t.Run("AssertRedirect", func(t *testing.T) {
		resp := NewMockResponse()
		resp.WriteHeader(http.StatusFound)
		resp.Header().Set("Location", "/dashboard")
		resp.AssertRedirect(t, "/dashboard")
	})
	
	t.Run("AssertBodyContains", func(t *testing.T) {
		resp := NewMockResponse()
		resp.Write([]byte("Hello, World!"))
		resp.AssertBodyContains(t, "World")
	})
	
	t.Run("AssertHeader", func(t *testing.T) {
		resp := NewMockResponse()
		resp.Header().Set("X-Custom-Header", "test-value")
		resp.AssertHeader(t, "X-Custom-Header", "test-value")
	})
	
	t.Run("AssertJSON", func(t *testing.T) {
		resp := NewMockResponse()
		resp.Header().Set("Content-Type", "application/json")
		resp.Write([]byte(`{"name":"test","value":42}`))
		
		var data map[string]interface{}
		resp.AssertJSON(t, &data)
		
		AssertEqual(t, "test", data["name"])
		AssertEqual(t, float64(42), data["value"]) // JSON numbers are float64
	})
}

func TestMockHTTPClient(t *testing.T) {
	t.Run("Single response", func(t *testing.T) {
		client := NewMockHTTPClient(MockHTTPResponse{
			StatusCode: 200,
			Body:       "test response",
			Headers:    map[string]string{"Content-Type": "text/plain"},
		})
		
		resp, err := client.Get("http://example.com/test")
		AssertNoError(t, err)
		AssertEqual(t, 200, resp.StatusCode)
		
		// Check request was recorded
		client.AssertRequestCount(t, 1)
		client.AssertRequestMade(t, "GET", "http://example.com/test")
	})
	
	t.Run("Multiple responses", func(t *testing.T) {
		client := NewMockHTTPClient(
			MockHTTPResponse{StatusCode: 200, Body: "first"},
			MockHTTPResponse{StatusCode: 201, Body: "second"},
			MockHTTPResponse{StatusCode: 404, Body: "not found"},
		)
		
		// First request
		resp1, err := client.Get("http://example.com/1")
		AssertNoError(t, err)
		AssertEqual(t, 200, resp1.StatusCode)
		
		// Second request
		resp2, err := client.Post("http://example.com/2", "application/json", nil)
		AssertNoError(t, err)
		AssertEqual(t, 201, resp2.StatusCode)
		
		// Third request
		resp3, err := client.Get("http://example.com/3")
		AssertNoError(t, err)
		AssertEqual(t, 404, resp3.StatusCode)
		
		// Verify requests
		client.AssertRequestCount(t, 3)
		client.AssertRequestMade(t, "GET", "http://example.com/1")
		client.AssertRequestMade(t, "POST", "http://example.com/2")
		client.AssertRequestMade(t, "GET", "http://example.com/3")
	})
	
	t.Run("Error response", func(t *testing.T) {
		expectedErr := errors.New("network error")
		client := NewMockHTTPClient(MockHTTPResponse{
			Error: expectedErr,
		})
		
		_, err := client.Get("http://example.com/error")
		AssertError(t, err)
		AssertEqual(t, expectedErr, err)
	})
}