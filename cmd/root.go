package cmd

func Execute() {
	// Setup flags and parse them
	setupFlagsAndParse()

	// handle management commands first
	if handled := dispatchSystemCommands(); handled {
		return
	}

	// Create logger
	logger := newLogger()

	// Load chain configuration if enabled
	hops := loadChainIfEnabled(logger)

	// Build base dialer (direct/tor)
	base := buildBaseDialer(logger)

	// Build the plan (with or without chaining)
	plan := buildPlan(logger, base, hops)

	// Run the server
	err := runServer(logger, plan)

	// Cleanup (e.g., stop Tor)
	cleanup()

	// Check for errors and log fatal if any
	if err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
