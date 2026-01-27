package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"proxychan/internal/models"
	"proxychan/internal/system"

	"golang.org/x/term"
)

func runAddUser(db *sql.DB) {
	username := prompt("Username")
	pass1 := promptPassword("Password")
	pass2 := promptPassword("Confirm password")

	if pass1 != pass2 {
		fatal(
			models.NewCLIError(
				"USER_PASS_MISMATCH",
				models.ExitUsage,
				"passwords do not match",
			),
		)
	}

	if err := system.AddUser(db, username, pass1); err != nil {
		fatal(
			models.
				Wrap(
					"USER_ADD_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to add user %q", username),
					err,
				),
		)
	}

	fmt.Println("user added:", username)
}

func runListUsers(db *sql.DB) {
	users, err := system.ListUsers(db)
	if err != nil {
		fatal(
			models.
				Wrap(
					"USER_LIST_FAIL",
					models.ExitRuntime,
					"failed to list users",
					err,
				),
		)
	}

	if len(users) == 0 {
		fmt.Println("no users defined")
		return
	}

	for _, u := range users {
		status, err := system.ListUserByUsername(db, u)
		if err != nil {
			fatal(
				models.
					Wrap(
						"USER_STATUS_FAIL",
						models.ExitRuntime,
						fmt.Sprintf("failed to read status for user %q", u),
						err,
					),
			)
		}
		fmt.Printf("- %s (%s)\n", u, status)
	}
}

func runListUser(db *sql.DB, username string) {
	status, err := system.ListUserByUsername(db, username)
	if err != nil {
		fatal(
			models.
				Wrap(
					"USER_LOOKUP_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to lookup user %q", username),
					err,
				),
		)
	}

	fmt.Printf("User %s is %s\n", username, status)
}

func runDeleteUser(db *sql.DB, username string) {
	if err := system.DeleteUser(db, username); err != nil {
		fatal(
			models.
				Wrap(
					"USER_DELETE_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to delete user %q", username),
					err,
				),
		)
	}

	fmt.Printf("User %s deleted.\n", username)
}

func runActivateUser(db *sql.DB, username string) {
	if err := system.ActivateUser(db, username); err != nil {
		fatal(
			models.
				Wrap(
					"USER_ACTIVATE_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to activate user %q", username),
					err,
				),
		)
	}

	fmt.Printf("User %s activated.\n", username)
}

func runDeactivateUser(db *sql.DB, username string) {
	if err := system.DeactivateUser(db, username); err != nil {
		fatal(
			models.
				Wrap(
					"USER_DEACTIVATE_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to deactivate user %q", username),
					err,
				),
		)
	}

	fmt.Printf("User %s deactivated.\n", username)
}

func runActivateAllUsers(db *sql.DB) {
	if err := system.ActivateAllUsers(db); err != nil {
		fatal(
			models.
				Wrap(
					"USER_ACTIVATE_ALL_FAIL",
					models.ExitRuntime,
					"failed to activate all users",
					err,
				),
		)
	}

	fmt.Println("All users activated.")
}

func runDeactivateAllUsers(db *sql.DB) {
	if err := system.DeactivateAllUsers(db); err != nil {
		fatal(
			models.
				Wrap(
					"USER_DEACTIVATE_ALL_FAIL",
					models.ExitRuntime,
					"failed to deactivate all users",
					err,
				),
		)
	}

	fmt.Println("All users deactivated.")
}

func prompt(label string) string {
	fmt.Print(label + ": ")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		fatal(
			models.
				Wrap(
					"INPUT_READ_FAIL",
					models.ExitIO,
					"failed to read input",
					err,
				),
		)
	}
	return strings.TrimSpace(text)
}

func promptPassword(label string) string {
	fmt.Print(label + ": ")
	bytePwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fatal(
			models.
				Wrap(
					"PASSWORD_READ_FAIL",
					models.ExitIO,
					"failed to read password",
					err,
				),
		)
	}
	return string(bytePwd)
}
