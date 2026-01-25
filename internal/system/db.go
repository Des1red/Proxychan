package system

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

func DBPath() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return "/var/lib/proxychan/proxychan.db", nil
	case "darwin":
		return "/Library/Application Support/ProxyChan/proxychan.db", nil
	case "windows":
		pd := os.Getenv("ProgramData")
		if pd == "" {
			return "", fmt.Errorf("ProgramData not set")
		}
		return filepath.Join(pd, "ProxyChan", "proxychan.db"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
func InitDB() (*sql.DB, error) {

	dbPath, err := DBPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
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
		active INTEGER NOT NULL DEFAULT 1,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS whitelist (
		cidr TEXT PRIMARY KEY,
		enabled INTEGER NOT NULL DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS whitelist_meta (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		version INTEGER NOT NULL
	);

	INSERT OR IGNORE INTO whitelist_meta (id, version)
	VALUES (1, 1);

	-- default allow localhost (IPv4 + IPv6)
	INSERT OR IGNORE INTO whitelist (cidr, enabled) VALUES
		('127.0.0.1/32', 1),
		('::1/128', 1);
	`

	_, err := db.Exec(schema)
	return err
}
