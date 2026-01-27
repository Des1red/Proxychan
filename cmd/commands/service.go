package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"proxychan/internal/models"
	systemserviceinstall "proxychan/internal/systemServiceInstall"
)

func currentBinaryPath() string {
	p, err := os.Executable()
	if err != nil {
		fmt.Println("failed to determine binary path:", err)
		os.Exit(1)
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		fmt.Println("failed to resolve binary path:", err)
		os.Exit(1)
	}
	return p
}

func runInstallService(cfg models.FlagConfig) {
	// Check if the user is running as root
	if os.Geteuid() != 0 {
		fmt.Println("Error: Installing the service requires root privileges. Please run with sudo.")
		os.Exit(1)
	}
	bin := currentBinaryPath()

	installCfg := systemserviceinstall.InstallConfig{
		BinaryPath: bin,
		ListenAddr: cfg.ListenAddr,
		HttpListen: cfg.HttpListen,
		Mode:       cfg.Mode,
		User:       os.Getenv("USER"),
		NoAuth:     cfg.NoAuth,
	}

	if err := systemserviceinstall.Install(installCfg); err != nil {
		fmt.Println("service installation failed:", err)
		os.Exit(1)
	}

	fmt.Println("service installed successfully")
}

func runRemoveService() {
	if err := systemserviceinstall.Remove(); err != nil {
		fmt.Println("service removal failed:", err)
		os.Exit(1)
	}
	fmt.Println("service removed successfully")
}
