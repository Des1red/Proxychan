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
		fatal(
			models.
				Wrap(
					"BIN_PATH_FAIL",
					models.ExitIO,
					"failed to determine binary path",
					err,
				),
		)
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		fatal(
			models.
				Wrap(
					"BIN_PATH_RESOLVE_FAIL",
					models.ExitIO,
					"failed to resolve binary path",
					err,
				),
		)
	}
	return p
}

func runInstallService(cfg models.FlagConfig) {
	// Check if the user is running as root
	if os.Geteuid() != 0 {
		fatal(
			models.
				NewCLIError(
					"SERVICE_NEEDS_ROOT",
					models.ExitUsage,
					"installing the service requires root privileges",
				).
				WithHint("run the command with sudo"),
		)
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
		fatal(
			models.
				Wrap(
					"SERVICE_INSTALL_FAIL",
					models.ExitExternal,
					"service installation failed",
					err,
				),
		)
	}

	fmt.Println("service installed successfully")
}

func runRemoveService() {
	if err := systemserviceinstall.Remove(); err != nil {
		fatal(
			models.
				Wrap(
					"SERVICE_REMOVE_FAIL",
					models.ExitExternal,
					"service removal failed",
					err,
				),
		)
	}
	fmt.Println("service removed successfully")
}
