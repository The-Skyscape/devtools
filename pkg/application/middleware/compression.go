package middleware

import (
	"cmp"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Compression provides gzip compression middleware
type Compression struct {
	level int
}

// NewCompression creates a new compression middleware
func NewCompression(level int) *Compression {
	level = cmp.Or(level, gzip.DefaultCompression)
	return &Compression{level}
}

// Handle returns the HTTP handler that compresses responses
func (c *Compression) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for small responses or already compressed content
		if r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Set headers before writing
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length") // Content-Length is not valid with compression

		// Create gzip writer
		gz, err := gzip.NewWriterLevel(w, c.level)
		if err != nil {
			// Fall back to no compression if there's an error
			next.ServeHTTP(w, r)
			return
		}

		// Wrap the response writer
		gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}

		// Process request
		next.ServeHTTP(gzw, r)
		
		// Close the gzip writer after processing
		gz.Close()
	})
}

// gzipResponseWriter wraps http.ResponseWriter to provide gzip compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Flush implements http.Flusher
func (w *gzipResponseWriter) Flush() {
	// Flush the gzip writer
	if gzw, ok := w.Writer.(*gzip.Writer); ok {
		gzw.Flush()
	}
	// Then flush the underlying response writer
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
