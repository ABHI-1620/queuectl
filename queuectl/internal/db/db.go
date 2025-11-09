package db

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"os"
)

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(".", 0755); err != nil {
		return nil, err
	}
	dsn := path + "?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}
