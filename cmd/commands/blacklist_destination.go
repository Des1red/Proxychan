package commands

import (
	"database/sql"
	"fmt"
	"os"
	"proxychan/internal/models"
	"proxychan/internal/system"
)

// block-destination
func runBlockDestination(db *sql.DB, target string) {
	if err := system.DenyDestination(db, target); err != nil {
		fatal(
			models.
				Wrap("DEST_BLOCK_FAIL", models.ExitRuntime,
					fmt.Sprintf("failed to block destination %q", target),
					err).
				WithHint("check destination format (IP, CIDR, domain, or .domain)"),
		)
	}
	fmt.Printf("destination blocked: %s\n", target)
}

// allow-destination
func runAllowDestination(db *sql.DB, target string) {
	if err := system.AllowDestination(db, target); err != nil {
		fatal(
			models.
				Wrap(
					"DEST_ALLOW_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to allow destination %q", target),
					err,
				),
		)
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
		fatal(
			models.
				Wrap(
					"DEST_LIST_FAIL",
					models.ExitRuntime,
					"failed to list destination blacklist",
					err,
				),
		)
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
		fatal(
			models.
				Wrap(
					"DEST_CLEAR_FAIL",
					models.ExitRuntime,
					"failed to clear destination blacklist",
					err,
				),
		)
	}

	if err := system.BumpDenylistVersion(db); err != nil {
		fatal(
			models.
				Wrap(
					"DEST_VERSION_FAIL",
					models.ExitRuntime,
					"failed to bump destination blacklist version",
					err,
				),
		)
	}

	fmt.Println("DESTINATION BLACKLIST CLEARED")
	fmt.Println("ALL DESTINATIONS ARE NOW ALLOWED")
}
