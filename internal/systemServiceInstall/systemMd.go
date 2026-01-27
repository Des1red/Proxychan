package systemserviceinstall

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const systemdUnit = `[Unit]
Description=ProxyChan SOCKS5 Proxy
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=%s
Restart=on-failure
RestartSec=3
KillSignal=SIGTERM
TimeoutStopSec=10

[Install]
WantedBy=multi-user.target
`

func installSystemd(binary string, args []string) error {
	unitPath := "/etc/systemd/system/proxychan.service"

	execStart := strings.Join(
		append([]string{binary}, args...),
		" ",
	)

	content := fmt.Sprintf(systemdUnit, execStart)

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
