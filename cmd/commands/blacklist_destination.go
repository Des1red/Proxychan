package commands

import (
	"database/sql"
	"fmt"
	"os"
	"proxychan/internal/system"
)

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
