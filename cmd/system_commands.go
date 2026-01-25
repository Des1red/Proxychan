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

func runListWhitelist(db *sql.DB) {
	entries, err := system.ListWhitelist(db)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("source whitelist is empty")
		return
	}

	fmt.Println("SOURCE WHITELIST")
	fmt.Println("----------------------------------------------")
	for _, e := range entries {
		state := "DISABLED"
		if e.Enabled {
			state = "ENABLED"
		}
		fmt.Printf("[%s] %s\n", state, e.CIDR)
	}
}

func runDeleteIP(db *sql.DB, ip string) {
	if err := system.DeleteIP(db, ip); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("deleted:", ip)
}

func runWhitelistStatus(db *sql.DB) {
	s, err := system.GetWhitelistStatus(db)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Printf(
		"version=%d total=%d enabled=%d disabled=%d\n",
		s.Version,
		s.Total,
		s.Enabled,
		s.Disabled,
	)
}

func runClearWhitelist(db *sql.DB) {
	if err := system.ClearWhitelist(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("WHITELIST CLEARED")
	fmt.Println("localhost entries preserved (127.0.0.1/32, ::1/128)")
}

// block-destination
func runBlockDestination(db *sql.DB, target string) {
	if err := system.DenyDestination(db, target); err != nil {
		fmt.Fprintf(os.Stderr, "failed to block destination %q: %v\n", target, err)
		os.Exit(1)
	}
	fmt.Printf("destination blocked: %s\n", target)
}

// allow-destination
func runAllowDestination(db *sql.DB, target string) {
	if err := system.AllowDestination(db, target); err != nil {
		fmt.Fprintf(os.Stderr, "failed to allow destination %q: %v\n", target, err)
		os.Exit(1)
	}
	fmt.Printf("destination allowed: %s\n", target)
}

// delete-destination
func runDeleteDestination(db *sql.DB, target string) {
	if err := system.DeleteDestination(db, target); err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete destination rule %q: %v\n", target, err)
		os.Exit(1)
	}
	fmt.Printf("destination rule deleted: %s\n", target)
}

// list-blacklist
func runListBlacklist(db *sql.DB) {
	rules, err := system.ListDenylist(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list blacklist: %v\n", err)
		os.Exit(1)
	}

	if len(rules) == 0 {
		fmt.Println("destination blacklist is empty")
		return
	}

	fmt.Println("DESTINATION BLACKLIST")
	fmt.Println("----------------------------------------------")
	for _, r := range rules {
		state := "DISABLED"
		if r.Enabled {
			state = "ENABLED"
		}
		fmt.Printf("[%s] %-14s %s\n", state, r.Type, r.Pattern)
	}
}

func runClearBlacklist(db *sql.DB) {
	if err := system.ClearDenylist(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := system.BumpDenylistVersion(db); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("DESTINATION BLACKLIST CLEARED")
	fmt.Println("ALL DESTINATIONS ARE NOW ALLOWED")
}
