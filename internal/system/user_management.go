package system

import "database/sql"

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
