package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"proxychan/internal/dialer"
	"proxychan/internal/logging"
	"proxychan/internal/server"
	"proxychan/internal/service"
	"proxychan/internal/socks5"
	"proxychan/internal/system"
	"syscall"
)

func mustInitDB() *sql.DB {
	db, err := system.InitDB()
	if err != nil {
		fmt.Println("db error:", err)
		os.Exit(1)
	}
	return db
}

// loadChainIfEnabled loads the chain configuration if dynamic chain is enabled.
func loadChainIfEnabled() []dialer.ChainHop {
	if !*dynamicChain {
		return nil
	}

	cfg, err := dialer.LoadChainConfig(*chainConfig)
	if err != nil {
		logging.GetLogger().Fatalf("failed to load chain config: %v", err)
	}
	return cfg.Chain
}

// buildBaseDialer selects the base dialer based on the mode.
func buildBaseDialer() dialer.Dialer {
	var d dialer.Dialer
	switch *mode {
	case "direct":
		d = dialer.NewDirect(*connectTimeout)
	case "tor":
		service.TorServiceStart(*torSocksAddr)
		d = socks5.NewTorSOCKS5(*torSocksAddr, *connectTimeout)
	default:
		logging.GetLogger().Fatalf("invalid -mode: %q (use direct|tor)", *mode)
	}
	return d
}

// buildPlan creates the dial plan based on the base dialer and chain configuration.
func buildPlan(base dialer.Dialer, hops []dialer.ChainHop) *dialer.Plan {
	var plan *dialer.Plan
	var err error

	if hops != nil {
		plan, err = dialer.NewChainedPlan(hops, base)
	} else {
		plan, err = dialer.NewPlan(base)
	}

	if err != nil {
		logging.GetLogger().Fatalf("dial plan error: %v", err)
	}
	return plan
}

// runServer starts the server with the given configuration.
func runServer(
	plan *dialer.Plan,
	authFn func(username, password string) error, db *sql.DB) error {
	requireAuth := server.RequiresAuth(*listenAddr)

	// Explicit override
	if *noAuth {
		requireAuth = false
		logging.GetLogger().Warn("authentication disabled via --no-auth")
	}

	srv := server.New(server.Config{
		ListenAddr:  *listenAddr,
		Dialer:      plan,
		IdleTimeout: *idleTimeout,
		Logger:      logging.GetLogger(), // Pass logrus logger
		RequireAuth: requireAuth,
		AuthFunc:    authFn,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return srv.Run(ctx, db)
}

// cleanup performs cleanup tasks (like stopping Tor service).
func cleanup() {
	service.TorServiceStop()
}
