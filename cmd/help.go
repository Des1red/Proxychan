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
		clihelp.F("del-user", "string", "Deletes existing user"),
		clihelp.F("list-users", "", "Prints list of existing users"),
		clihelp.F("list-user", "string", "Prints info of specific user"),
		clihelp.F("activate-user", "string", "Activates access to specific user"),
		clihelp.F("activate-all", "", "Activates access to specific user"),
		clihelp.F("deactivate-user", "string", "Deactivates access to specific user"),
		clihelp.F("deactivate-all", "", "Deactivates access to all users"),
	)
	fmt.Println()

	fmt.Println("White List management:")
	clihelp.Print(
		clihelp.F(
			"allow-ip",
			"string",
			"Allow IP or CIDR range (e.g. 192.168.1.5 or 192.168.1.0/24)",
		),
		clihelp.F(
			"block-ip",
			"string",
			"Disable access for an IP or CIDR range (keeps entry)",
		),
		clihelp.F(
			"delete-ip",
			"string",
			"Remove IP or CIDR range from whitelist entirely",
		),
		clihelp.F(
			"list-whitelist",
			"",
			"Print all whitelisted IPs and CIDR ranges",
		),
		clihelp.F(
			"status-whitelist",
			"",
			"Show whitelist version and entry count",
		),
		clihelp.F(
			"clear-whitelist",
			"",
			"Remove all whitelist entries and reset version",
		),
	)

	fmt.Println()
	fmt.Println("Doctor:")
	clihelp.Print(clihelp.F("doctor", "", "Prints Log and DB paths"))
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
