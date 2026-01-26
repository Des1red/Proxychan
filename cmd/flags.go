package cmd

import (
	"database/sql"
	"flag"
	"fmt"
	"net"
	"os"
	"proxychan/internal/dialer"
	"proxychan/internal/logging"
	"proxychan/internal/system"
	"time"
)

var (
	listenAddr     = flag.String("listen", "127.0.0.1:1080", "listen address for SOCKS5 proxy")
	httpListen     = flag.String("http-listen", "", "listen address for HTTP CONNECT proxy")
	mode           = flag.String("mode", "direct", "egress mode: direct | tor")
	torSocksAddr   = flag.String("tor-socks", "127.0.0.1:9050", "Tor SOCKS5 address (mode=tor)")
	connectTimeout = flag.Duration("connect-timeout", 10*time.Second, "outbound connect timeout")
	idleTimeout    = flag.Duration("idle-timeout", 2*time.Minute, "idle timeout for tunnels (0 disables)")
	noAuth         = flag.Bool("no-auth", false, "disable username/password authentication (IP whitelist still enforced)")

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

	// tor-socks misuse check
	if *mode != "tor" {
		const defaultTor = "127.0.0.1:9050"

		if *torSocksAddr != "" && *torSocksAddr != defaultTor {
			return false, "--tor-socks can only be used with --mode tor"
		}
	}

	// --no-auth is meaningless on localhost
	if *noAuth {
		host, _, err := net.SplitHostPort(*listenAddr)
		if err != nil {
			return false, "invalid listen address"
		}
		ip := net.ParseIP(host)
		if ip != nil && ip.IsLoopback() {
			return false, "--no-auth is unnecessary when binding to localhost"
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
	case "allow-ip":
		if len(args) != 2 {
			fmt.Println("usage: proxychan allow-ip <IP>")
			os.Exit(1)
		}
		runAllowIP(db, args[1])
	case "block-ip":
		if len(args) != 2 {
			fmt.Println("usage: proxychan block-ip <IP>")
			os.Exit(1)
		}
		runBlockIP(db, args[1])
	case "del-ip":
		if len(args) != 2 {
			fmt.Println("usage: proxychan delete-ip <IP>")
			os.Exit(1)
		}
		runDeleteIP(db, args[1])
		return true

	case "list-whitelist":
		runListWhitelist(db)
		return true

	case "clear-whitelist":
		runClearWhitelist(db)
		return true

	case "status-whitelist":
		runWhitelistStatus(db)
		return true

	case "block-dest":
		if len(args) != 2 {
			fmt.Println("usage: proxychan block-dest<ip|cidr|domain|.domain>")
			os.Exit(1)
		}
		runBlockDestination(db, args[1])
		return true

	case "allow-dest":
		if len(args) != 2 {
			fmt.Println("usage: proxychan allow-dest<ip|cidr|domain|.domain>")
			os.Exit(1)
		}
		runAllowDestination(db, args[1])
		return true

	case "del-dest":
		if len(args) != 2 {
			fmt.Println("usage: proxychan delete-dest <ip|cidr|domain|.domain>")
			os.Exit(1)
		}
		runDeleteDestination(db, args[1])
		return true

	case "list-blacklist":
		runListBlacklist(db)
		return true

	case "list-connections":
		runListConnections(db)
		return true

	case "doctor":
		dbPath, _ := system.DBPath()
		logDir, _ := logging.LogDir()
		runDoctor(dbPath, logDir)
		return true

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

	// If a system command is present, skip flag validation
	if len(flag.Args()) > 0 {
		return
	}

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
