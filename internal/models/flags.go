package models

import (
	"fmt"
	"reflect"
	"time"
)

type FlagConfig struct {
	ListenAddr     string        `flag:"listen"`
	HttpListen     string        `flag:"http-listen" omitEmpty:"true"`
	Mode           string        `flag:"mode"`
	TorSocksAddr   string        `flag:"tor-socks" omitEmpty:"true"`
	ConnectTimeout time.Duration `flag:"connect-timeout"`
	IdleTimeout    time.Duration `flag:"idle-timeout"`
	NoAuth         bool          `flag:"no-auth"`
	DynamicChain   bool          `flag:"dynamic-chain"`
	ChainConfig    string        `flag:"chain-config" omitEmpty:"true"`
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

func (cfg FlagConfig) ToArgs() ([]string, error) {
	var args []string

	v := reflect.ValueOf(cfg)
	t := reflect.TypeOf(cfg)

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fieldVal := v.Field(i)

		flagName := fieldType.Tag.Get("flag")
		if flagName == "" {
			continue // explicitly ignored
		}

		omitEmpty := fieldType.Tag.Get("omitEmpty") == "true"
		flag := "--" + flagName

		switch fieldVal.Kind() {

		case reflect.Bool:
			if fieldVal.Bool() {
				args = append(args, flag)
			}

		case reflect.String:
			val := fieldVal.String()
			if val == "" && omitEmpty {
				continue
			}
			args = append(args, flag, val)

		case reflect.Int64:
			// time.Duration is int64 underneath
			if fieldType.Type != reflect.TypeOf(time.Duration(0)) {
				return nil, fmt.Errorf("unsupported int64 field: %s", fieldType.Name)
			}

			d := time.Duration(fieldVal.Int())
			if d == 0 && omitEmpty {
				continue
			}
			args = append(args, flag, d.String())

		default:
			return nil, fmt.Errorf(
				"unsupported field type %s for %s",
				fieldVal.Kind(),
				fieldType.Name,
			)
		}
	}

	return args, nil
}
