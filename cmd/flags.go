package cmd

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"proxychan/internal/dialer"
	"proxychan/internal/logging"
	"time"
)

var (
	listenAddr     = flag.String("listen", "127.0.0.1:1080", "listen address for SOCKS5 proxy")
	mode           = flag.String("mode", "direct", "egress mode: direct | tor")
	torSocksAddr   = flag.String("tor-socks", "127.0.0.1:9050", "Tor SOCKS5 address (mode=tor)")
	connectTimeout = flag.Duration("connect-timeout", 10*time.Second, "outbound connect timeout")
	idleTimeout    = flag.Duration("idle-timeout", 2*time.Minute, "idle timeout for tunnels (0 disables)")

	dynamicChain = flag.Bool("dynamic-chain", false, "enable dynamic SOCKS5 hop chaining from YAML config")
	chainConfig  = flag.String("chain-config", "", "path to YAML chain config (required when -dynamic-chain=true)")
)

func badFlagUse() (bool, string) {
	if *dynamicChain {
		if *chainConfig == "" {
			return false, "-chain-config is required when -dynamic-chain is enabled"
		}

		if _, err := dialer.LoadChainConfig(*chainConfig); err != nil {
			return false, err.Error()
		}
	}

	return true, ""
}

func dispatchSystemCommands(db *sql.DB) bool {
	args := flag.Args()
	if len(args) == 0 {
		return false
	}

	switch args[0] {
	case "add-user":
		runAddUser(db)

	case "list-users":
		runListUsers(db)

	case "list-user":
		if len(args) != 2 {
			fmt.Println("usage: proxychan list-user <username>")
			os.Exit(1)
		}
		runListUser(db, args[1])

	case "del-user":
		if len(args) != 2 {
			fmt.Println("usage: proxychan delete-user <username>")
			os.Exit(1)
		}
		runDeleteUser(db, args[1])

	case "activate-user":
		if len(args) != 2 {
			fmt.Println("usage: proxychan activate-user <username>")
			os.Exit(1)
		}
		runActivateUser(db, args[1])

	case "deactivate-user":
		if len(args) != 2 {
			fmt.Println("usage: proxychan deactivate-user <username>")
			os.Exit(1)
		}
		runDeactivateUser(db, args[1])

	case "activate-all":
		runActivateAllUsers(db)

	case "deactivate-all":
		runDeactivateAllUsers(db)

	case "install-service":
		runInstallService()

	case "remove-service":
		runRemoveService()

	default:
		fmt.Printf("unknown command: %s\n\n", args[0])
		printHelp()
		os.Exit(1)
	}

	return true
}

// setupFlagsAndParse sets up the command-line flags and parses them.
func setupFlagsAndParse() {
	flag.Usage = printHelp
	flag.Parse()

	// Check flag usage validity
	ok, msg := badFlagUse()
	if !ok {
		// Log the error with logrus for flag validation issues
		logging.GetLogger().Errorf("Flag validation error: %s", msg)

		// Display the message and exit
		fmt.Println(msg)
		printHelp()
		os.Exit(1)
	}
}
