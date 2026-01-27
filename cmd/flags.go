package cmd

import (
	"fmt"
	"net"
	"os"
	"proxychan/cmd/commands"
	"proxychan/internal/dialer"
	"proxychan/internal/logging"
	"proxychan/internal/models"

	"github.com/spf13/pflag"
)

var cfg = models.DefaultFlagConfig

func defineFlags() {
	pflag.CommandLine.SortFlags = false

	pflag.StringVar(
		&cfg.ListenAddr,
		"listen",
		cfg.ListenAddr,
		"listen address for SOCKS5 proxy",
	)

	pflag.StringVar(
		&cfg.HttpListen,
		"http-listen",
		cfg.HttpListen,
		"listen address for HTTP CONNECT proxy",
	)

	pflag.StringVar(
		&cfg.Mode,
		"mode",
		cfg.Mode,
		"egress mode: direct | tor",
	)

	pflag.StringVar(
		&cfg.TorSocksAddr,
		"tor-socks",
		cfg.TorSocksAddr,
		"Tor SOCKS5 address (mode=tor)",
	)

	pflag.DurationVar(
		&cfg.ConnectTimeout,
		"connect-timeout",
		cfg.ConnectTimeout,
		"outbound connect timeout",
	)

	pflag.DurationVar(
		&cfg.IdleTimeout,
		"idle-timeout",
		cfg.IdleTimeout,
		"idle timeout for tunnels (0 disables)",
	)

	pflag.BoolVar(
		&cfg.NoAuth,
		"no-auth",
		cfg.NoAuth,
		"disable username/password authentication (IP whitelist still enforced)",
	)

	pflag.BoolVar(
		&cfg.DynamicChain,
		"dynamic-chain",
		cfg.DynamicChain,
		"enable dynamic SOCKS5 hop chaining from YAML config",
	)

	pflag.StringVar(
		&cfg.ChainConfig,
		"chain-config",
		cfg.ChainConfig,
		"path to YAML chain config (required when -dynamic-chain=true)",
	)
}

func badFlagUse(cfg models.FlagConfig) (bool, string) {
	if cfg.DynamicChain {
		if cfg.ChainConfig == "" {
			return false, "-chain-config is required when -dynamic-chain is enabled"
		}

		if _, err := dialer.LoadChainConfig(cfg.ChainConfig); err != nil {
			return false, err.Error()
		}
	}

	// tor-socks misuse check
	if cfg.Mode != "tor" {
		const defaultTor = "127.0.0.1:9050"

		if cfg.TorSocksAddr != "" && cfg.TorSocksAddr != defaultTor {
			return false, "--tor-socks can only be used with --mode tor"
		}
	}

	// --no-auth is meaningless on localhost
	if cfg.NoAuth {
		host, _, err := net.SplitHostPort(cfg.ListenAddr)
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

// setupFlagsAndParse parses flags and validates runtime usage
func setupFlagsAndParse() {
	defineFlags()
	pflag.Usage = commands.PrintHelp
	pflag.Parse()

	// Check flag usage validity
	ok, msg := badFlagUse(cfg)
	if !ok {
		// Log the error with logrus for flag validation issues
		logging.GetLogger().Errorf("Flag validation error: %s", msg)

		// Display the message and exit
		fmt.Println(msg)
		commands.PrintHelp()
		os.Exit(1)
	}
}
