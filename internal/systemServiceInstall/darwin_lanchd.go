package systemserviceinstall

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const launchdPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.proxychan.proxy</string>
  <key>ProgramArguments</key>
  <array>
%s
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
</dict>
</plist>
`

func installLaunchd(cfg InstallConfig) error {
	const plistPath = "/Library/LaunchDaemons/com.proxychan.proxy.plist"

	args := []string{
		fmt.Sprintf("    <string>%s</string>", cfg.BinaryPath),
		"    <string>-listen</string>",
		fmt.Sprintf("    <string>%s</string>", cfg.ListenAddr),
		"    <string>-mode</string>",
		fmt.Sprintf("    <string>%s</string>", cfg.Mode),
	}

	if cfg.NoAuth {
		args = append(args, "    <string>--no-auth</string>")
	}
	if cfg.HttpListen != "" {
		args = append(args, fmt.Sprintf("    <string>%s</string>", cfg.HttpListen))
	}
	content := fmt.Sprintf(
		launchdPlist,
		strings.Join(args, "\n"),
	)

	if err := os.WriteFile(plistPath, []byte(content), 0644); err != nil {
		return err
	}

	cmds := [][]string{
		{"launchctl", "bootout", "system/com.proxychan.proxy"},
		{"launchctl", "bootstrap", "system", plistPath},
	}

	for _, c := range cmds {
		_ = exec.Command(c[0], c[1:]...).Run()
	}

	return nil
}

func removeLaunchd() error {
	const plistPath = "/Library/LaunchDaemons/com.proxychan.proxy.plist"

	// Unload (best-effort)
	_ = exec.Command("launchctl", "bootout", "system/com.proxychan.proxy").Run()

	// Remove plist
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
