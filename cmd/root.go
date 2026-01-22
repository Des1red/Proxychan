package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"proxychan/internal/dialer"
	"proxychan/internal/server"
	"proxychan/internal/service"
	"syscall"
)

func Execute() {
	flag.Usage = printHelp
	flag.Parse()

	ok, msg := badFlagUse()
	if !ok {
		fmt.Println(msg)
		printHelp()
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "[proxychan] ", log.LstdFlags|log.Lmicroseconds)

	// --- load dynamic chain (if enabled) ---
	var hops []dialer.ChainHop
	if *dynamicChain {
		cfg, err := dialer.LoadChainConfig(*chainConfig)
		if err != nil {
			logger.Fatalf("failed to load chain config: %v", err)
		}
		hops = cfg.Chain
	}

	// --- select base dialer ---
	var d dialer.Dialer
	switch *mode {
	case "direct":
		d = dialer.NewDirect(*connectTimeout)
	case "tor":
		service.TorServiceStart(*torSocksAddr)
		d = dialer.NewTorSOCKS5(*torSocksAddr, *connectTimeout)
	default:
		logger.Fatalf("invalid -mode: %q (use direct|tor)", *mode)
	}

	// --- select plan ---
	var plan *dialer.Plan
	var err error

	if *dynamicChain {
		plan, err = dialer.NewChainedPlan(hops, d)
	} else {
		plan, err = dialer.NewPlan(d)
	}

	if err != nil {
		logger.Fatalf("dial plan error: %v", err)
	}

	// --- server ---
	srv := server.New(server.Config{
		ListenAddr:  *listenAddr,
		Dialer:      plan,
		IdleTimeout: *idleTimeout,
		Logger:      logger,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = srv.Run(ctx)

	service.TorServiceStop()

	if err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
