package database

import (
	"database/sql"
	"fmt"
)

var (
	ErrIterStop = fmt.Errorf("stop iteration")
)

type Iter struct {
	Conn *sql.DB
	Text string
	Args []any
}

type reader func(ScanFunc) error
type ScanFunc func(...any) error

func (i *Iter) Exec() error {
	_, err := i.Conn.Exec(i.Text, i.Args...)
	return err
}

func (i *Iter) Scan(args ...any) error {
	row := i.Conn.QueryRow(i.Text, i.Args...)
	if err := row.Err(); err != nil {
		return err
	}

	return row.Scan(args...)
}

func (i *Iter) All(fn reader) error {
	rows, err := i.Conn.Query(i.Text, i.Args...)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		if err := fn(rows.Scan); err != nil {
			return err
		}
	}

	return nil
}

func (i *Iter) Page(limit int, fn reader) (more bool, err error) {
	rows, err := i.Conn.Query(i.Text, i.Args...)
	if err != nil {
		return false, err
	}

	defer rows.Close()
	var count int
	for rows.Next() {
		if count++; count == limit+1 {
			return true, nil
		}

		if err := fn(rows.Scan); err != nil {
			return false, err
		}
	}

	return false, nil
}
