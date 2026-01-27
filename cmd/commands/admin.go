package commands

import (
	"database/sql"
	"fmt"
	"proxychan/internal/models"
	"proxychan/internal/system"
)

func runSetAdminPassword(db *sql.DB) {
	exists, err := adminPasswordExists(db)
	if err != nil {
		fatal(
			models.Wrap(
				"ADMIN_PWD_CHECK_FAIL",
				models.ExitExternal,
				"failed to check admin password status",
				err,
			),
		)
	}

	var current *string
	if exists {
		pwd := promptPassword("Current admin password")
		current = &pwd
	}

	newPwd := promptPassword("New admin password")
	confirm := promptPassword("Confirm admin password")

	if newPwd != confirm {
		fatal(
			models.Wrap(
				"ADMIN_PWD_MISMATCH",
				models.ExitUsage,
				"passwords do not match",
				nil,
			),
		)
	}

	if err := system.SetOrRotateAdminPassword(db, current, newPwd); err != nil {
		fatal(
			models.Wrap(
				"ADMIN_PWD_SET_FAIL",
				models.ExitExternal,
				"failed to set admin password",
				err,
			),
		)
	}

	fmt.Println("admin password updated")
}

func adminPasswordExists(db *sql.DB) (bool, error) {
	var dummy int
	err := db.QueryRow(
		`SELECT 1 FROM admin_auth WHERE id = 1`,
	).Scan(&dummy)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
