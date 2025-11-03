package libsql

import (
	"database/sql"
	"log"
	"path/filepath"
	"time"

	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/tursodatabase/go-libsql"
)

type LibSQL struct {
	*sql.DB
}

func Open(name, url, token string) *LibSQL {
	path := filepath.Join(database.DataDir(), name)

	db, err := libsql.NewEmbeddedReplicaConnector(path, url,
		libsql.WithSyncInterval(time.Second*10),
		libsql.WithAuthToken(token))

	if err != nil {
		log.Fatal("Failed to replicate to remote db:", err)
	}

	if _, err = db.Sync(); err != nil {
		log.Fatal("Failed to sync to remote db:", err)
	}

	return &LibSQL{DB: sql.OpenDB(db)}
}

func (db *LibSQL) Model() database.Model {
	return database.Model{DB: db}
}

func (db *LibSQL) NewModel(id string) database.Model {
	return database.Model{DB: db, ID: id, CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func (db *LibSQL) Query(query string, args ...any) *database.Iter {
	return &database.Iter{Conn: db.DB, Text: query, Args: args}
}

func (db *LibSQL) Dynamic() *database.DynamicDB {
	return database.Dynamic(db)
}
