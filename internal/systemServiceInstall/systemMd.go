package systemserviceinstall

import (
	"fmt"
	"os"
	"os/exec"
)

const systemdUnit = `[Unit]
Description=ProxyChan SOCKS5 Proxy
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=%s -listen %s -mode %s
Restart=on-failure
RestartSec=3
KillSignal=SIGTERM
TimeoutStopSec=10

[Install]
WantedBy=multi-user.target
`

func installSystemd(cfg InstallConfig) error {
	unitPath := "/etc/systemd/system/proxychan.service"

	content := fmt.Sprintf(
		systemdUnit,
		cfg.BinaryPath,
		cfg.ListenAddr,
		cfg.Mode,
	)

	if err := os.WriteFile(unitPath, []byte(content), 0644); err != nil {
		return err
	}

	steps := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", "proxychan"},
		{"systemctl", "restart", "proxychan"},
	}

	for _, cmd := range steps {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return err
		}
	}

	return nil
}

func removeSystemd() error {
	unitPath := "/etc/systemd/system/proxychan.service"

	// Stop + disable (best-effort, idempotent)
	_ = exec.Command("systemctl", "stop", "proxychan").Run()
	_ = exec.Command("systemctl", "disable", "proxychan").Run()

	// Remove unit file
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return err
	}

	return nil
}
