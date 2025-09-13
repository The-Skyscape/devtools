// Package mocks provides mock implementations of DevTools interfaces for testing.
//
// All interfaces in DevTools that are designed for external implementation
// have corresponding mock implementations in this package. This ensures that
// applications using DevTools can easily test their code without real dependencies.
//
// Usage:
//
//	// Create a mock database
//	db := mocks.NewMockDatabase()
//	db.QueryFunc = func(query string, args ...any) *database.Iter {
//	    // Return test data
//	}
//
//	// Create a mock secrets store
//	secrets := mocks.NewMockSecrets()
//	secrets.StoreSecret("test/path", map[string]any{"key": "value"})
//
//	// Create a mock middleware
//	middleware := mocks.NewMockMiddleware()
//	middleware.Block() // Block all requests
//
// Available Mocks:
//
//	- MockDatabase: database.Database interface
//	- MockSecrets: security.Secrets interface
//	- MockHandler: application.Handler interface
//	- MockMiddleware: application.Middleware interface
//	- MockHost: containers.Host interface (in containers package)
//	- MockPlatform: hosting.Platform interface (in hosting package)
//	- MockProvider: payments.Provider interface (in payments package)
//	- MockEmailProvider: emailing.Provider interface (in emailing package)
package mocks

// This file serves as documentation and ensures the package is not empty