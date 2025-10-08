package remote

import (
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/engines/libsql"
)

func Database(name, url, token string) *database.DynamicDB {
	return libsql.Open(name, url, token).Dynamic()
}
