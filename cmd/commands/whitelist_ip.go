package commands

import (
	"database/sql"
	"fmt"
	"os"
	"proxychan/internal/system"
)

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
