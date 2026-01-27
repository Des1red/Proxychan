package system

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

// ListUsers returns all users in the database
func ListUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT username FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ListUserByUsername returns the status (active/inactive) of a specific user
func ListUserByUsername(db *sql.DB, username string) (string, error) {
	// Check if the user exists and get their active status
	var active int
	err := db.QueryRow(
		`SELECT active FROM user_status
		JOIN users ON users.id = user_status.user_id
		WHERE users.username = ?`,
		username,
	).Scan(&active)

	if err == sql.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}

	// Determine the user's status
	status := "active"
	if active == 0 {
		status = "inactive"
	}
	return status, nil
}

func DeleteUser(db *sql.DB, username string) error {
	res, err := db.Exec(
		`DELETE FROM users WHERE username = ?`,
		username,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrUserNotFound
	}

	return nil
}

func AddUser(db *sql.DB, username, password string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRow(
		`SELECT 1 FROM users WHERE username = ?`,
		username,
	).Scan(&exists)

	if err == nil {
		return ErrUserExists
	}
	if err != sql.ErrNoRows {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	res, err := tx.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		username,
		hash,
	)
	if err != nil {
		return err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO user_status (user_id, active) VALUES (?, 0)`,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func Authenticate(db *sql.DB, username, password string) error {
	var hash string
	err := db.QueryRow(
		`SELECT password_hash FROM users WHERE username = ?`,
		username,
	).Scan(&hash)

	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	); err != nil {
		return ErrBadCredential
	}

	return nil
}

func IsActive(db *sql.DB, username string) (bool, error) {
	var active int
	err := db.QueryRow(
		`SELECT active FROM user_status
		JOIN users ON users.id = user_status.user_id
		WHERE users.username = ?`,
		username,
	).Scan(&active)

	if err == sql.ErrNoRows {
		return false, ErrUserNotFound
	}
	if err != nil {
		return false, err
	}

	return active == 1, nil
}

func ActivateUser(db *sql.DB, username string) error {
	_, err := db.Exec(
		`UPDATE user_status 
		SET active = 1 
		WHERE user_id = (SELECT id FROM users WHERE username = ?)`,
		username,
	)
	if err != nil {
		return err
	}
	return nil
}

func DeactivateUser(db *sql.DB, username string) error {
	_, err := db.Exec(
		`UPDATE user_status 
		SET active = 0 
		WHERE user_id = (SELECT id FROM users WHERE username = ?)`,
		username,
	)
	if err != nil {
		return err
	}
	return nil
}

// ActivateAllUsers activates all users in the database.
func ActivateAllUsers(db *sql.DB) error {
	_, err := db.Exec(
		`UPDATE user_status SET active = 1`,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeactivateAllUsers deactivates all users in the database.
func DeactivateAllUsers(db *sql.DB) error {
	_, err := db.Exec(
		`UPDATE user_status SET active = 0`,
	)
	if err != nil {
		return err
	}
	return nil
}
