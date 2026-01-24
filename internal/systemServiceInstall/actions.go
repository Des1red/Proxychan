package systemserviceinstall

import (
	"fmt"
	"runtime"
)

type InstallConfig struct {
	BinaryPath string
	ListenAddr string
	Mode       string // direct | tor
	User       string // optional
}

func Install(cfg InstallConfig) error {
	switch runtime.GOOS {
	case "linux":
		return installSystemd(cfg)
	case "darwin":
		return installLaunchd(cfg)
	case "windows":
		return installWindowsService(cfg)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func Remove() error {
	switch runtime.GOOS {
	case "linux":
		return removeSystemd()
	case "darwin":
		return removeLaunchd()
	case "windows":
		return removeWindowsService()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
