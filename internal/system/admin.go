package system

import (
	"database/sql"
	"errors"
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
