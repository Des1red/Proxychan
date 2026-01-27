package commands

import (
	"database/sql"
	"fmt"

	"proxychan/internal/models"
	"proxychan/internal/system"
)

func runAllowIP(db *sql.DB, ip string) {
	if err := system.AllowIP(db, ip); err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_ALLOW_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to allow IP %q", ip),
					err,
				),
		)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_VERSION_FAIL",
					models.ExitRuntime,
					"failed to bump whitelist version",
					err,
				),
		)
	}

	fmt.Println("allowed:", ip)
}

func runBlockIP(db *sql.DB, ip string) {
	if err := system.BlockIP(db, ip); err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_BLOCK_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to block IP %q", ip),
					err,
				),
		)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_VERSION_FAIL",
					models.ExitRuntime,
					"failed to bump whitelist version",
					err,
				),
		)
	}

	fmt.Println("blocked:", ip)
}

func runDeleteIP(db *sql.DB, ip string) {
	if err := system.DeleteIP(db, ip); err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_DELETE_FAIL",
					models.ExitRuntime,
					fmt.Sprintf("failed to delete IP %q", ip),
					err,
				),
		)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_VERSION_FAIL",
					models.ExitRuntime,
					"failed to bump whitelist version",
					err,
				),
		)
	}

	fmt.Println("deleted:", ip)
}

func runListWhitelist(db *sql.DB) {
	entries, err := system.ListWhitelist(db)
	if err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_LIST_FAIL",
					models.ExitRuntime,
					"failed to list source whitelist",
					err,
				),
		)
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

func runWhitelistStatus(db *sql.DB) {
	s, err := system.GetWhitelistStatus(db)
	if err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_STATUS_FAIL",
					models.ExitRuntime,
					"failed to read whitelist status",
					err,
				),
		)
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
		fatal(
			models.
				Wrap(
					"WHITELIST_CLEAR_FAIL",
					models.ExitRuntime,
					"failed to clear whitelist",
					err,
				),
		)
	}

	if err := system.BumpWhitelistVersion(db); err != nil {
		fatal(
			models.
				Wrap(
					"WHITELIST_VERSION_FAIL",
					models.ExitRuntime,
					"failed to bump whitelist version",
					err,
				),
		)
	}

	fmt.Println("WHITELIST CLEARED")
	fmt.Println("localhost entries preserved (127.0.0.1/32, ::1/128)")
}
