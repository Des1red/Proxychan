package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"proxychan/internal/dialer"
	"proxychan/internal/server"
	"proxychan/internal/service"
	"proxychan/internal/socks5"
	"syscall"
)

// newLogger creates and returns a new logger.
func newLogger() *log.Logger {
	return log.New(os.Stdout, "[proxychan] ", log.LstdFlags|log.Lmicroseconds)
}

// loadChainIfEnabled loads the chain configuration if dynamic chain is enabled.
func loadChainIfEnabled(logger *log.Logger) []dialer.ChainHop {
	if !*dynamicChain {
		return nil
	}

	cfg, err := dialer.LoadChainConfig(*chainConfig)
	if err != nil {
		logger.Fatalf("failed to load chain config: %v", err)
	}
	return cfg.Chain
}

// buildBaseDialer selects the base dialer based on the mode.
func buildBaseDialer(logger *log.Logger) dialer.Dialer {
	var d dialer.Dialer
	switch *mode {
	case "direct":
		d = dialer.NewDirect(*connectTimeout)
	case "tor":
		service.TorServiceStart(*torSocksAddr)
		d = socks5.NewTorSOCKS5(*torSocksAddr, *connectTimeout)
	default:
		logger.Fatalf("invalid -mode: %q (use direct|tor)", *mode)
	}
	return d
}

// buildPlan creates the dial plan based on the base dialer and chain configuration.
func buildPlan(logger *log.Logger, base dialer.Dialer, hops []dialer.ChainHop) *dialer.Plan {
	var plan *dialer.Plan
	var err error

	if hops != nil {
		plan, err = dialer.NewChainedPlan(hops, base)
	} else {
		plan, err = dialer.NewPlan(base)
	}

	if err != nil {
		logger.Fatalf("dial plan error: %v", err)
	}
	return plan
}

// runServer starts the server with the given configuration.
func runServer(logger *log.Logger, plan *dialer.Plan) error {
	requireAuth := server.RequiresAuth(*listenAddr)
	srv := server.New(server.Config{
		ListenAddr:  *listenAddr,
		Dialer:      plan,
		IdleTimeout: *idleTimeout,
		Logger:      logger,
		RequireAuth: requireAuth,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return srv.Run(ctx)
}

// cleanup performs cleanup tasks (like stopping Tor service).
func cleanup() {
	service.TorServiceStop()
}
