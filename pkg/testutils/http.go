package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// MockRequest creates a new HTTP request for testing
func MockRequest(method, path string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, path, body)
	return req
}

// MockJSONRequest creates a new HTTP request with JSON body
func MockJSONRequest(method, path string, body interface{}) *http.Request {
	jsonBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// MockFormRequest creates a new HTTP request with form data
func MockFormRequest(method, path string, values url.Values) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// MockResponse wraps httptest.ResponseRecorder with helper methods
type MockResponse struct {
	*httptest.ResponseRecorder
}

// NewMockResponse creates a new mock response recorder
func NewMockResponse() *MockResponse {
	return &MockResponse{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

// AssertStatus checks if the response has the expected status code
func (r *MockResponse) AssertStatus(t *testing.T, expected int) {
	t.Helper()
	if r.Code != expected {
		t.Errorf("Expected status %d, got %d", expected, r.Code)
	}
}

// AssertOK checks if the response is 200 OK
func (r *MockResponse) AssertOK(t *testing.T) {
	t.Helper()
	r.AssertStatus(t, http.StatusOK)
}

// AssertRedirect checks if the response is a redirect
func (r *MockResponse) AssertRedirect(t *testing.T, expectedLocation string) {
	t.Helper()
	if r.Code != http.StatusFound && r.Code != http.StatusSeeOther && r.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected redirect status, got %d", r.Code)
	}
	
	location := r.Header().Get("Location")
	if expectedLocation != "" && location != expectedLocation {
		t.Errorf("Expected redirect to %s, got %s", expectedLocation, location)
	}
}

// AssertJSON checks if the response is valid JSON and unmarshals it
func (r *MockResponse) AssertJSON(t *testing.T, target interface{}) {
	t.Helper()
	
	contentType := r.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}
	
	if err := json.Unmarshal(r.Body.Bytes(), target); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}
}

// AssertBodyContains checks if the response body contains a string
func (r *MockResponse) AssertBodyContains(t *testing.T, expected string) {
	t.Helper()
	body := r.Body.String()
	if !strings.Contains(body, expected) {
		t.Errorf("Response body does not contain expected string: %s", expected)
	}
}

// AssertHeader checks if a header has the expected value
func (r *MockResponse) AssertHeader(t *testing.T, key, expected string) {
	t.Helper()
	actual := r.Header().Get(key)
	if actual != expected {
		t.Errorf("Expected header %s: %s, got %s", key, expected, actual)
	}
}

// MockHTTPClient creates a mock HTTP client with custom transport
type MockHTTPClient struct {
	*http.Client
	Requests  []*http.Request
	Responses []MockHTTPResponse
	index     int
}

// MockHTTPResponse defines a mock response
type MockHTTPResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	Error      error
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient(responses ...MockHTTPResponse) *MockHTTPClient {
	client := &MockHTTPClient{
		Requests:  make([]*http.Request, 0),
		Responses: responses,
	}
	
	client.Client = &http.Client{
		Transport: client,
	}
	
	return client
}

// RoundTrip implements http.RoundTripper
func (c *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	c.Requests = append(c.Requests, req)
	
	if c.index >= len(c.Responses) {
		return nil, io.EOF
	}
	
	resp := c.Responses[c.index]
	c.index++
	
	if resp.Error != nil {
		return nil, resp.Error
	}
	
	httpResp := &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(strings.NewReader(resp.Body)),
		Header:     make(http.Header),
		Request:    req,
	}
	
	for k, v := range resp.Headers {
		httpResp.Header.Set(k, v)
	}
	
	return httpResp, nil
}

// AssertRequestMade checks if a request was made to the expected URL
func (c *MockHTTPClient) AssertRequestMade(t *testing.T, method, url string) {
	t.Helper()
	
	for _, req := range c.Requests {
		if req.Method == method && req.URL.String() == url {
			return
		}
	}
	
	t.Errorf("Expected request %s %s was not made", method, url)
}

// AssertRequestCount checks the number of requests made
func (c *MockHTTPClient) AssertRequestCount(t *testing.T, expected int) {
	t.Helper()
	
	if len(c.Requests) != expected {
		t.Errorf("Expected %d requests, got %d", expected, len(c.Requests))
	}
}