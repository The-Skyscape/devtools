package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"github.com/The-Skyscape/devtools/pkg/database"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite3 struct {
	*sql.DB
	name, root string
}

func Open(name string, tables fs.FS) *SQLite3 {
	db := SQLite3{name: name, root: database.DataDir()}
	err := os.MkdirAll(db.root, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	dbFilePath := filepath.Join(db.root, name)
	if db.DB, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL", dbFilePath)); err != nil {
		log.Fatalf("Failed to connect to datatabase: %v", err)
	}

	if err := db.DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if _, err = db.DB.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		log.Fatalf("Failed to set WAL mode: %v", err)
	}

	if tables != nil {
		var fs source.Driver
		if fs, err = iofs.New(tables, "tables"); err != nil {
			log.Fatalf("Failed to create migration driver: %v", err)
		}

		var m *migrate.Migrate
		dest := fmt.Sprintf("sqlite3://%s/%s", db.root, db.name)
		if m, err = migrate.NewWithSourceInstance("iofs", fs, dest); err != nil {
			log.Fatalf("Failed to create migration instance: %v", err)
		}

		if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Failed to migrate database: %v", err)
		}
	}

	return &db
}

func (db *SQLite3) Model() database.Model {
	return database.Model{DB: db}
}

func (db *SQLite3) NewModel(id string) database.Model {
	return database.Model{DB: db, ID: id, CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func (db *SQLite3) Query(query string, args ...any) *database.Iter {
	return &database.Iter{Conn: db.DB, Text: query, Args: args}
}

func (db *SQLite3) Dynamic() *database.DynamicDB {
	return database.Dynamic(db)
}
