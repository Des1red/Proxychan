package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type torController interface {
	IsTorRunning() error
	StartTor() error
	StopTor() error
}

// ---- shared runner ----

func runCmd(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}

func runCmdInteractive(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// ---- Linux ----

type linuxTor struct{}

func (l *linuxTor) IsTorRunning() error {
	return runCmd("systemctl", "is-active", "--quiet", "tor")
}

func (l *linuxTor) StartTor() error {
	return runCmdInteractive("systemctl", "start", "tor")
}

func (l *linuxTor) StopTor() error {
	return runCmdInteractive("systemctl", "stop", "tor")
}

// ---- macOS (Homebrew) ----

type darwinTor struct{}

func (d *darwinTor) IsTorRunning() error {
	return runCmd("brew", "services", "list")
}

func (d *darwinTor) StartTor() error {
	return runCmdInteractive("brew", "services", "start", "tor")
}

func (d *darwinTor) StopTor() error {
	return runCmdInteractive("brew", "services", "stop", "tor")
}

// ---- Windows (SCM) ----

type windowsTor struct{}

func (w *windowsTor) IsTorRunning() error {
	return runCmd("sc", "query", "tor")
}

func (w *windowsTor) StartTor() error {
	return runCmdInteractive("sc", "start", "tor")
}

func (w *windowsTor) StopTor() error {
	return runCmdInteractive("sc", "stop", "tor")
}

// ---- selector ----

func getTorController() (torController, error) {
	switch runtime.GOOS {
	case "linux":
		return &linuxTor{}, nil
	case "darwin":
		return &darwinTor{}, nil
	case "windows":
		return &windowsTor{}, nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
