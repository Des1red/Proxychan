package cmd

import "proxychan/internal/logging"

// Execute runs the main execution flow
func Execute() {
	// Setup flags and parse them
	setupFlagsAndParse()

	// Setup structured logging
	logging.SetupLogger()

	// Handle management commands first (like add-user, del-user)
	if handled := dispatchSystemCommands(); handled {
		return
	}

	// Load chain configuration if dynamic chain is enabled
	hops := loadChainIfEnabled()

	// Build base dialer (direct/tor)
	base := buildBaseDialer()

	// Build the dial plan
	plan := buildPlan(base, hops)

	// Run the server
	err := runServer(plan)

	// Cleanup (stop Tor service if needed)
	cleanup()

	// Check for errors and log fatal if any
	if err != nil {
		logging.GetLogger().Fatalf("server error: %v", err)
	}
}
