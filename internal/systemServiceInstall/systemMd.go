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
User=%s
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
		cfg.User,
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
