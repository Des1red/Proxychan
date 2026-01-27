package systemserviceinstall

import (
	"fmt"
	"os/exec"
	"strings"
)

func installWindowsService(binary string, args []string) error {
	const serviceName = "ProxyChan"

	fullCmd := strings.Join(
		append([]string{binary}, args...),
		" ",
	)

	// Best-effort stop + delete (idempotent)
	_ = exec.Command("sc.exe", "stop", serviceName).Run()
	_ = exec.Command("sc.exe", "delete", serviceName).Run()

	createCmd := exec.Command(
		"sc.exe",
		"create", serviceName,
		"binPath=", fullCmd,
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

func removeWindowsService() error {
	const serviceName = "ProxyChan"

	// Best-effort stop + delete
	_ = exec.Command("sc.exe", "stop", serviceName).Run()
	_ = exec.Command("sc.exe", "delete", serviceName).Run()

	return nil
}
