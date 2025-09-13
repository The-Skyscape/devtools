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

// Open creates and returns a new SQLite3 database connection.
// For backward compatibility, this function still uses log.Fatal on errors.
// New code should use OpenWithError instead.
func Open(name string, tables fs.FS) *SQLite3 {
	db, err := OpenWithError(name, tables)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return db
}

// OpenWithError creates and returns a new SQLite3 database connection,
// returning an error instead of calling log.Fatal on failure.
func OpenWithError(name string, tables fs.FS) (*SQLite3, error) {
	db := SQLite3{name: name, root: database.DataDir()}
	err := os.MkdirAll(db.root, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dbFilePath := filepath.Join(db.root, name)
	if db.DB, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL", dbFilePath)); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.DB.Ping(); err != nil {
		db.DB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err = db.DB.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		db.DB.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if tables != nil {
		var fs source.Driver
		if fs, err = iofs.New(tables, "tables"); err != nil {
			db.DB.Close()
			return nil, fmt.Errorf("failed to create migration driver: %w", err)
		}

		var m *migrate.Migrate
		dest := fmt.Sprintf("sqlite3://%s/%s", db.root, db.name)
		if m, err = migrate.NewWithSourceInstance("iofs", fs, dest); err != nil {
			db.DB.Close()
			return nil, fmt.Errorf("failed to create migration instance: %w", err)
		}

		if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			db.DB.Close()
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return &db, nil
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
