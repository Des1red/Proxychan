package cmd

import (
	"fmt"

	"github.com/Des1red/clihelp"
)

func printHelp() {
	fmt.Println("proxychan - Minimal SOCKS5 proxy with optional chaining")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  proxychan [flags]")
	fmt.Println()

	// ─── Core ────────────────────────────────────────────────
	fmt.Println("Core:")
	clihelp.Print(
		clihelp.F("--listen", "address", "Listen address for SOCKS5 proxy"),
		clihelp.F("--mode", "string", "Egress mode (direct | tor)"),
		clihelp.F("--connect-timeout", "duration", "Outbound connect timeout"),
		clihelp.F("--idle-timeout", "duration", "Idle tunnel timeout (0 disables)"),
	)
	fmt.Println()

	// ─── Tor ─────────────────────────────────────────────────
	fmt.Println("Tor:")
	clihelp.Print(
		clihelp.F("--tor-socks", "address", "Tor SOCKS5 address"),
	)
	fmt.Println()

	// ─── Chaining ────────────────────────────────────────────
	fmt.Println("Chaining:")
	clihelp.Print(
		clihelp.F("--dynamic-chain", "", "Enable dynamic SOCKS5 hop chaining"),
		clihelp.F("--chain-config", "path", "YAML chain config (required if enabled)"),
	)
	fmt.Println()

	fmt.Println("User management:")
	clihelp.Print(
		clihelp.F("add-user", "", "Initiates user creation"),
		clihelp.F("list-users", "", "Prints list of existing users"),
		clihelp.F("del-user", "string", "Deletes existing user"),
		clihelp.F("activate-user", "string", "Activates access to specific user"),
		clihelp.F("activate-all", "string", "Activates access to specific user"),
		clihelp.F("deactivate-user", "string", "Deactivates access to specific user"),
		clihelp.F("deactivate-all", "string", "Deactivates access to all users"),
	)
	fmt.Println()

	fmt.Println("auto-configuration:")
	clihelp.Print(
		clihelp.F("install-service", "", "Install proxychan as a system service"),
		clihelp.F("remove-service", "", "Remove proxychan system service"),
	)
	fmt.Println()
	// ─── Notes ───────────────────────────────────────────────
	fmt.Println("Notes:")
	fmt.Println("  • SOCKS5 hops only")
	fmt.Println("  • No retries")
	fmt.Println("  • Dead hop = hard failure")
	fmt.Println("  • Tor must be either base OR hop, not both")

	fmt.Println("Auto-config Notes:")
	fmt.Println("  • install-service / remove-service require root/admin privileges")
	fmt.Println("  • use a compiled binary (not `go run`)")
	fmt.Println("  • usage : sudo ./proxychan --flag1 --flag2 install-service")
}
