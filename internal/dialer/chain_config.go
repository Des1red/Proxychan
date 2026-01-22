package dialer

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ChainConfig struct {
	Chain []ChainHop `yaml:"chain"`
}

type ChainHop struct {
	Type string `yaml:"type"`
	Addr string `yaml:"addr"`
}

func LoadChainConfig(path string) (*ChainConfig, error) {
	if path == "" {
		return nil, errors.New("chain config path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read chain config: %w", err)
	}

	var cfg ChainConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse chain config yaml: %w", err)
	}

	if len(cfg.Chain) == 0 {
		return nil, errors.New("chain config: empty chain")
	}

	for i, hop := range cfg.Chain {
		if hop.Type != "socks5" {
			return nil, fmt.Errorf("chain hop %d: unsupported type %q", i, hop.Type)
		}
		if hop.Addr == "" {
			return nil, fmt.Errorf("chain hop %d: empty addr", i)
		}
	}

	return &cfg, nil
}
