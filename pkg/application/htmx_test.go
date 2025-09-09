package application

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsHTMX(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{
			name:     "HTMX request",
			headers:  map[string]string{"HX-Request": "true"},
			expected: true,
		},
		{
			name:     "Non-HTMX request",
			headers:  map[string]string{},
			expected: false,
		},
		{
			name:     "HTMX with wrong value",
			headers:  map[string]string{"HX-Request": "false"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			if got := IsHTMX(req); got != tt.expected {
				t.Errorf("IsHTMX() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHTMXHeaders(t *testing.T) {
	t.Run("HTMXRedirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		HTMXRedirect(w, "/new-path")
		
		if got := w.Header().Get("HX-Redirect"); got != "/new-path" {
			t.Errorf("HX-Redirect header = %v, want /new-path", got)
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("Status code = %v, want %v", w.Code, http.StatusNoContent)
		}
	})

	t.Run("HTMXRefresh", func(t *testing.T) {
		w := httptest.NewRecorder()
		HTMXRefresh(w)
		
		if got := w.Header().Get("HX-Refresh"); got != "true" {
			t.Errorf("HX-Refresh header = %v, want true", got)
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("Status code = %v, want %v", w.Code, http.StatusNoContent)
		}
	})

	t.Run("HTMXTrigger", func(t *testing.T) {
		w := httptest.NewRecorder()
		HTMXTrigger(w, "event1", "event2")
		
		if got := w.Header().Get("HX-Trigger"); got != "event1,event2" {
			t.Errorf("HX-Trigger header = %v, want event1,event2", got)
		}
	})
}

func TestCurrentURL(t *testing.T) {
	t.Run("From HTMX header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api", nil)
		req.Header.Set("HX-Current-URL", "https://example.com/page")
		
		if got := CurrentURL(req); got != "https://example.com/page" {
			t.Errorf("CurrentURL() = %v, want https://example.com/page", got)
		}
	})

	t.Run("From request URL", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api", nil)
		
		if got := CurrentURL(req); got != "/api" {
			t.Errorf("CurrentURL() = %v, want /api", got)
		}
	})
}