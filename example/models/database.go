package models

import (
	"github.com/The-Skyscape/devtools/pkg/authentication"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// Global database and collections - initialized once at startup
var (
	// DB is the SQLite database for the application
	DB *database.DynamicDB

	// Auth provides user authentication (signin/signup/sessions)
	Auth *authentication.Collection

	// Ducks is our main model collection (CRUD operations)
	Ducks *database.Collection[*Duck]

	// Emails handles email sending with templates
	// Templates are loaded in main.go from embedded filesystem
	Emails *emailing.Collection
)

// InitDB sets up the database and all collections
func InitDB(db *database.DynamicDB) {
	// Store the database instance
	DB = db

	// Initialize Authentication package
	Auth = authentication.Manage(DB)

	// Initialize Ducks model collection
	Ducks = database.Manage(DB, new(Duck))

	// Initialize Emailing package
	Emails = emailing.Manage(DB,
		emailing.WithFrom("noreply@example.com", "Duck Debug"),
	)
}
