package libsql

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/tursodatabase/go-libsql"
)

type LibSQL struct {
	*sql.DB
	connector *libsql.Connector
}

func Open(name, url, token string) *LibSQL {
	path := filepath.Join(database.DataDir(), name)

	if info, err := os.Stat(path); err != nil || time.Since(info.ModTime()) > time.Hour {
		log.Println("Syncing database (missing or stale)...")
	}

	log.Println("Opening database:", name)
	db, err := libsql.NewEmbeddedReplicaConnector(path, url,
		libsql.WithSyncInterval(time.Second*30),
		libsql.WithAuthToken(token))

	if err != nil {
		log.Fatal("Failed to replicate to remote db:", err)
	}

	if info, err := os.Stat(path); err != nil || time.Since(info.ModTime()) > time.Hour {
		log.Println("Syncing database (missing or stale)...")
		if _, err := db.Sync(); err != nil {
			log.Fatal("Failed to sync to remote db:", err)
		}
	} else {
		log.Println("Using cached database, syncing in background...")
		go db.Sync()
	}

	return &LibSQL{DB: sql.OpenDB(db), connector: db}
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

func (db *LibSQL) Sync() error {
	if db.connector == nil {
		return nil
	}
	_, err := db.connector.Sync()
	return err
}
