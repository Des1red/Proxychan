package models

import "time"

type FlagConfig struct {
	ListenAddr     string
	HttpListen     string
	Mode           string
	TorSocksAddr   string
	ConnectTimeout time.Duration
	IdleTimeout    time.Duration
	NoAuth         bool
	DynamicChain   bool
	ChainConfig    string
}

var DefaultFlagConfig = FlagConfig{
	ListenAddr:     "127.0.0.1:1080",
	HttpListen:     "",
	Mode:           "direct",
	TorSocksAddr:   "127.0.0.1:9050",
	ConnectTimeout: 10 * time.Second,
	IdleTimeout:    2 * time.Minute,
	NoAuth:         false,
	DynamicChain:   false,
	ChainConfig:    "",
}
