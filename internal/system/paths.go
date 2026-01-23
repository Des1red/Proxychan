package system

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the proxychan config directory path,
// creating it if it does not already exist.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "proxychan")

	if err := os.MkdirAll(dir, 0o700); err != nil { // 0o700 private to the user
		return "", err
	}

	return dir, nil
}
