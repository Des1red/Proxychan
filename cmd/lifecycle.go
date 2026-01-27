package cmd

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"proxychan/cmd/commands"
	"proxychan/internal/dialer"
	"proxychan/internal/logging"
	"proxychan/internal/models"
	"proxychan/internal/server"
	"proxychan/internal/service"
	"proxychan/internal/socks5"
	"proxychan/internal/system"
	"syscall"
)

func mustInitDB() *sql.DB {
	db, err := system.InitDB()
	if err != nil {
		commands.Fatal(
			models.Wrap(
				"DB_INIT_FAIL",
				models.ExitIO,
				"failed to initialize database",
				err,
			),
		)
	}
	return db
}

// loadChainIfEnabled loads the chain configuration if dynamic chain is enabled.
func loadChainIfEnabled() []dialer.ChainHop {
	if !cfg.DynamicChain {
		return nil
	}

	chainCfg, err := dialer.LoadChainConfig(cfg.ChainConfig)

	if err != nil {
		logging.GetLogger().Fatalf("failed to load chain config: %v", err)
	}
	return chainCfg.Chain
}

// buildBaseDialer selects the base dialer based on the mode.
func buildBaseDialer() dialer.Dialer {
	var d dialer.Dialer
	switch cfg.Mode {
	case "direct":
		d = dialer.NewDirect(cfg.ConnectTimeout)
	case "tor":
		service.TorServiceStart(cfg.TorSocksAddr)
		d = socks5.NewTorSOCKS5(cfg.TorSocksAddr, cfg.ConnectTimeout)
	default:
		logging.GetLogger().Fatalf("invalid -mode: %q (use direct|tor)", cfg.Mode)
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
	requireAuth := server.RequiresAuth(cfg.ListenAddr)

	// Explicit override
	if cfg.NoAuth {
		requireAuth = false
		logging.GetLogger().Warn("authentication disabled via --no-auth")
	}

	srv := server.New(server.Config{
		ListenAddr:     cfg.ListenAddr,
		HTTPListenAddr: cfg.HttpListen,

		Dialer:      plan,
		IdleTimeout: cfg.IdleTimeout,
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
