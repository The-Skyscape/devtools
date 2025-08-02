package models

import (
	"github.com/The-Skyscape/devtools/pkg/authentication"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/local"
)

var (
	// DB is the example application's database
	DB = local.Database("example.db")

	// Auth is the DB's authentication collection
	Auth = authentication.Manage(DB)

	// Ducks is a collection of our Ducks model
	Ducks = database.Manage(DB, new(Duck))
)
