package cmd

import (
	"proxychan/cmd/commands"
	"proxychan/internal/logging"
	"proxychan/internal/system"
)

// Execute runs the main execution flow
func Execute() {
	// Setup structured logging
	logging.SetupLogger()
	// Setup flags and parse them
	setupFlagsAndParse()

	//Init db
	db := mustInitDB()
	defer db.Close()
	authFn := func(username, password string) error {
		return system.Authenticate(db, username, password)
	}
	// Handle management commands first (like add-user, del-user)
	if handled := commands.DispatchSystemCommands(db, cfg); handled {
		return
	}

	// Load chain configuration if dynamic chain is enabled
	hops := loadChainIfEnabled()

	// Build base dialer (direct/tor)
	base := buildBaseDialer()

	// Build the dial plan
	plan := buildPlan(base, hops)

	// Run the server
	err := runServer(plan, authFn, db)

	// Cleanup (stop Tor service if needed)
	cleanup()

	// Check for errors and log fatal if any
	if err != nil {
		logging.GetLogger().Fatalf("server error: %v", err)
	}
}
