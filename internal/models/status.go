package models

type RuntimeConfig struct {
	DisableTorOnExit bool
}

var DefaultRuntimeConfig = RuntimeConfig{
	DisableTorOnExit: false,
}
