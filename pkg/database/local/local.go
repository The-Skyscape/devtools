package local

import (
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/engines/sqlite3"
)

// Our local database engine is built ontop of sqlite3
// in the future we may add more options to configure
// what engine we want to be using and what modules
// we want to load.
func Database(name string) *database.DynamicDB {
	return sqlite3.Open(name, nil).Dynamic()
}
