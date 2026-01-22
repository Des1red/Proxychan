package cmd

import "fmt"

type helpFlag struct {
	name string
	arg  string
	desc string
}

func calcWidths(flags []helpFlag) (flagW, argW int) {
	for _, f := range flags {
		if len(f.name) > flagW {
			flagW = len(f.name)
		}
		if len(f.arg) > argW {
			argW = len(f.arg)
		}
	}
	return
}

func printFlags(flags []helpFlag) {
	flagW, argW := calcWidths(flags)
	for _, f := range flags {
		fmt.Printf(
			"  %-*s  %-*s  %s\n",
			flagW,
			f.name,
			argW,
			f.arg,
			f.desc,
		)
	}
}

func printHelp() {
	fmt.Println("proxychan - Minimal SOCKS5 proxy with optional chaining")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  proxychan [flags]\n")
	fmt.Println()

	// ─── Core ────────────────────────────────────────────────
	fmt.Println("Core:")
	printFlags([]helpFlag{
		{"--listen", "address", "Listen address for SOCKS5 proxy"},
		{"--mode", "string", "Egress mode (direct | tor)"},
		{"--connect-timeout", "duration", "Outbound connect timeout"},
		{"--idle-timeout", "duration", "Idle tunnel timeout (0 disables)"},
	})
	fmt.Println()

	// ─── Tor ─────────────────────────────────────────────────
	fmt.Println("Tor:")
	printFlags([]helpFlag{
		{"--tor-socks", "address", "Tor SOCKS5 address"},
	})
	fmt.Println()

	// ─── Chaining ────────────────────────────────────────────
	fmt.Println("Chaining:")
	printFlags([]helpFlag{
		{"--dynamic-chain", "", "Enable dynamic SOCKS5 hop chaining"},
		{"--chain-config", "path", "YAML chain config (required if enabled)"},
	})
	fmt.Println()

	// ─── Notes ───────────────────────────────────────────────
	fmt.Println("Notes:")
	fmt.Println("  • SOCKS5 hops only")
	fmt.Println("  • No retries")
	fmt.Println("  • Dead hop = hard failure")
	fmt.Println("  • Tor must be either base OR hop, not both")
}
