package database

import (
	"cmp"
	"fmt"
	"os"
	"time"
)

type Database interface {
	// modeling api
	Model() Model
	NewModel(id string) Model

	// querying api
	Query(string, ...any) *Iter
}

type Model struct {
	DB        Database
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Model) GetModel() *Model { return m }
func (m *Model) SetDB(d Database) { m.DB = d }

func DataDir() string {
	if os.Getenv("INTERNAL_DATA") != "" {
		return os.Getenv("INTERNAL_DATA")
	}
	root, _ := os.UserHomeDir()
	root = cmp.Or(root, "/tmp")
	data := fmt.Sprintf("%s/.skyscape", root)
	os.MkdirAll(data, os.ModePerm)
	return data
}
