package system

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

// SetOrRotateAdminPassword
// - current == nil  → initial setup
// - current != nil  → verify + rotate
func SetOrRotateAdminPassword(db *sql.DB, current *string, next string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var storedHash string
	err = tx.QueryRow(
		`SELECT password_hash FROM admin_auth WHERE id = 1`,
	).Scan(&storedHash)

	if err != nil {
		if err == sql.ErrNoRows {
			// No admin password exists
			if current != nil {
				return errors.New("admin password not set")
			}
			// initial insert
			hash, err := hashPassword(next)
			if err != nil {
				return err
			}
			_, err = tx.Exec(
				`INSERT INTO admin_auth (id, password_hash, updated_at)
				 VALUES (1, ?, ?)`,
				hash,
				time.Now(),
			)
			if err != nil {
				return err
			}
			return tx.Commit()
		}
		return err
	}

	// Admin password exists → rotation path
	if current == nil {
		return errors.New("admin password already set")
	}

	if !verifyPassword(storedHash, *current) {
		return errors.New("invalid admin password")
	}

	newHash, err := hashPassword(next)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE admin_auth
		 SET password_hash = ?, updated_at = ?
		 WHERE id = 1`,
		newHash,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func AdminPasswordConfigured(db *sql.DB) (bool, error) {
	var v int
	err := db.QueryRow(
		`SELECT 1 FROM admin_auth WHERE id = 1`,
	).Scan(&v)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func VerifyAdminCredentials(db *sql.DB, password string) error {
	var hash string
	err := db.QueryRow(
		`SELECT password_hash FROM admin_auth WHERE id = 1`,
	).Scan(&hash)
	if err != nil {
		return err
	}

	if !verifyPassword(hash, password) {
		return errors.New("invalid admin credentials")
	}
	return nil
}

var (
	internalSecretOnce sync.Once
	internalSecret     string
	internalSecretErr  error
)

func InternalAdminSecret() (string, error) {
	internalSecretOnce.Do(func() {
		internalSecret, internalSecretErr = loadOrCreateSecret("/var/lib/proxychan/admin_internal.secret")
	})
	return internalSecret, internalSecretErr
}

func loadOrCreateSecret(path string) (string, error) {
	// Try read first
	if b, err := os.ReadFile(path); err == nil {
		s := strings.TrimSpace(string(b))
		if s == "" {
			return "", errors.New("internal admin secret file is empty")
		}
		return s, nil
	}

	// Create new (root-only)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := hex.EncodeToString(b) + "\n"

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		// If it already exists (race), read it
		if b2, err2 := os.ReadFile(path); err2 == nil {
			ss := strings.TrimSpace(string(b2))
			if ss == "" {
				return "", errors.New("internal admin secret file is empty")
			}
			return ss, nil
		}
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(s); err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}
