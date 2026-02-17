package systemserviceinstall

import (
	"os"
	"os/exec"
	"strings"
)

func selinuxEnforcing() bool {
	data, err := os.ReadFile("/sys/fs/selinux/enforce")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "1"
}

func installSELinuxBinary(src string) (string, error) {
	dst := "/usr/local/bin/proxychan"

	// copy binary
	if err := exec.Command("cp", src, dst).Run(); err != nil {
		return "", err
	}

	if err := exec.Command("chmod", "755", dst).Run(); err != nil {
		return "", err
	}

	// fix label
	_ = exec.Command("restorecon", "-v", dst).Run()

	return dst, nil
}
