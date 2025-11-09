package db

import (
	"database/sql"
)

func Migrate(db *sql.DB) error {
	schema := `
	PRAGMA journal_mode=WAL;
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		command TEXT NOT NULL,
		state TEXT NOT NULL CHECK (state IN ('pending','processing','completed','failed','dead')),
		attempts INTEGER NOT NULL DEFAULT 0,
		max_retries INTEGER NOT NULL DEFAULT 3,
		base_backoff INTEGER NOT NULL DEFAULT 2,
		priority INTEGER NOT NULL DEFAULT 0,
		run_at DATETIME NOT NULL DEFAULT (datetime('now')),
		next_run_at DATETIME,
		lease_expires_at DATETIME,
		worker_id TEXT,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
		last_error TEXT
	);
	CREATE TABLE IF NOT EXISTS dlq (
		id TEXT PRIMARY KEY,
		original_id TEXT NOT NULL,
		command TEXT NOT NULL,
		attempts INTEGER NOT NULL,
		last_error TEXT,
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	INSERT OR IGNORE INTO config(key, value) VALUES
	  ('max_retries','3'),
	  ('backoff_base','2'),
	  ('lease_seconds','60'),
	  ('job_timeout_seconds','30'),
	  ('poll_interval_ms','500');
	`
	_, err := db.Exec(schema)
	return err
}
