package systemserviceinstall

import (
	"fmt"
	"runtime"
)

func Install(binary string, args []string) error {
	switch runtime.GOOS {
	case "linux":
		return installSystemd(binary, args)
	case "darwin":
		return installLaunchd(binary, args)
	case "windows":
		return installWindowsService(binary, args)
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
