package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"proxychan/internal/system"
	"strings"

	"golang.org/x/term"
)

func runAddUser(db *sql.DB) {
	username := prompt("Username")
	pass1 := promptPassword("Password")
	pass2 := promptPassword("Confirm password")

	if pass1 != pass2 {
		fmt.Println("passwords do not match")
		os.Exit(1)
	}

	if err := system.AddUser(db, username, pass1); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("user added:", username)
}

func runListUsers(db *sql.DB) {
	users, err := system.ListUsers(db)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if len(users) == 0 {
		fmt.Println("no users defined")
		return
	}

	for _, u := range users {
		active, err := system.ListUserByUsername(db, u) // Use new function to get user status
		if err != nil {
			fmt.Println("error checking user status:", err)
			os.Exit(1)
		}
		fmt.Printf("- %s (%s)\n", u, active)
	}
}

func runListUser(db *sql.DB, username string) {
	status, err := system.ListUserByUsername(db, username)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Printf("User %s is %s\n", username, status)
}

func runDeleteUser(db *sql.DB, username string) {
	if err := system.DeleteUser(db, username); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("User %s deleted.\n", username)
}

func prompt(label string) string {
	fmt.Print(label + ": ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func promptPassword(label string) string {
	fmt.Print(label + ": ")
	bytePwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Println("failed to read password")
		os.Exit(1)
	}
	return string(bytePwd)
}

func runActivateUser(db *sql.DB, username string) {
	if err := system.ActivateUser(db, username); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("User %s activated.\n", username)
}

func runDeactivateUser(db *sql.DB, username string) {
	if err := system.DeactivateUser(db, username); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("User %s deactivated.\n", username)
}

// runActivateAllUsers activates all users in the system.
func runActivateAllUsers(db *sql.DB) {
	if err := system.ActivateAllUsers(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	} else {
		fmt.Println("All users activated.")
	}
}

// runDeactivateAllUsers deactivates all users in the system.
func runDeactivateAllUsers(db *sql.DB) {
	if err := system.DeactivateAllUsers(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	} else {
		fmt.Println("All users deactivated.")
	}
}

func runAllowIP(db *sql.DB, ip string) {
	if err := system.AllowIP(db, ip); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("allowed:", ip)
}

func runBlockIP(db *sql.DB, ip string) {
	if err := system.BlockIP(db, ip); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("blocked:", ip)
}
