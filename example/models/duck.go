package models

import "github.com/The-Skyscape/devtools/pkg/application"

// Duck represents a rubber duck for debugging
// Demonstrates the Model pattern with embedded base
type Duck struct {
	application.Model            // Provides ID, CreatedAt, UpdatedAt
	Name        string           // Duck's name (required)
	Color       string           // Color: Yellow, Blue, Red, etc.
	Description string           // What this duck helps debug
	UserID      string           // Owner (would be from auth.CurrentUser)
}

// Table returns the database table name
// Required by the database package for all models
func (*Duck) Table() string { 
	return "ducks" 
}
