package systemserviceinstall

import (
	"fmt"
	"os/exec"
	"strings"
)

func installWindowsService(cfg InstallConfig) error {
	const serviceName = "ProxyChan"

	args := strings.Join([]string{
		cfg.BinaryPath,
		"-listen", cfg.ListenAddr,
		"-mode", cfg.Mode,
	}, " ")

	// Best-effort stop + delete (idempotent)
	_ = exec.Command("sc.exe", "stop", serviceName).Run()
	_ = exec.Command("sc.exe", "delete", serviceName).Run()

	createCmd := exec.Command(
		"sc.exe",
		"create", serviceName,
		"binPath=", args,
		"start=", "auto",
	)

	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := exec.Command("sc.exe", "start", serviceName).Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}
