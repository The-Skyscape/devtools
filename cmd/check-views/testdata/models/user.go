package models

import "github.com/The-Skyscape/devtools/pkg/application"

// User model for testing
type User struct {
	application.Model
	Email    string
	Name     string
	IsActive bool
}

// Table returns the table name
func (u *User) Table() string {
	return "users"
}

// GetInitials returns user initials (method)
func (u *User) GetInitials() string {
	// Simple implementation
	return "JD"
}
