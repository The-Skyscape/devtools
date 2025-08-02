package models

import "github.com/The-Skyscape/devtools/pkg/database/local"

var (
	// DB is the example application's database
	DB = local.Database("example.db")

	// Auth is the DB's authentication repository
	Auth = authentication.Manage(DB)

	Ducks = database.Manage(DB, new(Duck))
)
