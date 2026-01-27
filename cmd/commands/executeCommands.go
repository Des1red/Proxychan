package commands

import (
	"database/sql"
	"fmt"
	"os"
	"proxychan/internal/logging"
	"proxychan/internal/models"
	"proxychan/internal/system"

	"github.com/spf13/pflag"
)

func DispatchSystemCommands(db *sql.DB, cfg models.FlagConfig) bool {
	args := pflag.Args()
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
		runInstallService(cfg)

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
