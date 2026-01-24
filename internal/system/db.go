package system

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get user config dir: %w", err)
	}

	dir := filepath.Join(base, "proxychan")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create proxychan dir: %w", err)
	}

	dbPath := filepath.Join(dir, "proxychan.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Safety: fail early if DB is not writable
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func initSchema(db *sql.DB) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_status (
		user_id INTEGER PRIMARY KEY,
		blocked INTEGER NOT NULL DEFAULT 0,
		locked_until DATETIME,
		active INTEGER NOT NULL DEFAULT 1,  -- 1 for active, 0 for inactive
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	_, err := db.Exec(schema)
	return err
}
