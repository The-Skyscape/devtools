package models

import (
	"github.com/The-Skyscape/devtools/pkg/authentication"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/local"
	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// Global database and collections - initialized once at startup
var (
	// DB is the SQLite database for the application
	DB = local.Database("example.db")

	// Auth provides user authentication (signin/signup/sessions)
	Auth = authentication.Manage(DB)

	// Ducks is our main model collection (CRUD operations)
	Ducks = database.Manage(DB, new(Duck))

	// Emails handles email sending with templates
	// Templates are loaded in main.go from embedded filesystem
	Emails = emailing.Manage(DB,
		emailing.WithFrom("noreply@example.com", "Duck Debug"),
	)
)
