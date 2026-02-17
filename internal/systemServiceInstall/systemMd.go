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

// --- escape args so systemd parses them correctly ---
func systemdEscape(args []string) string {
	var escaped []string
	for _, a := range args {
		if strings.ContainsAny(a, " \t\"") {
			a = "\"" + strings.ReplaceAll(a, "\"", "\\\"") + "\""
		}
		escaped = append(escaped, a)
	}
	return strings.Join(escaped, " ")
}

func installSystemd(binary string, args []string) error {
	unitPath := "/etc/systemd/system/proxychan.service"

	// --- SELinux safe binary location ---
	if selinuxEnforcing() {
		newPath, err := installSELinuxBinary(binary)
		if err != nil {
			return fmt.Errorf("failed to install SELinux binary: %w", err)
		}
		binary = newPath

		// remember we installed it
		_ = os.MkdirAll("/var/lib/proxychan", 0755)
		_ = os.WriteFile("/var/lib/proxychan/selinux-installed", []byte("1"), 0644)
	}

	// ----------------------------------------
	execStart := systemdEscape(append([]string{binary}, args...))
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

	// Stop + disable (best-effort)
	_ = exec.Command("systemctl", "stop", "proxychan").Run()
	_ = exec.Command("systemctl", "disable", "proxychan").Run()

	// Remove SELinux-installed binary only if we created it
	if _, err := os.Stat("/var/lib/proxychan/selinux-installed"); err == nil {
		_ = os.Remove("/usr/local/bin/proxychan")
		_ = os.Remove("/var/lib/proxychan/selinux-installed")
	}

	// Remove unit file
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Reload + clear failure state
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return err
	}
	_ = exec.Command("systemctl", "reset-failed").Run()

	return nil
}
