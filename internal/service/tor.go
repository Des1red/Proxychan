package service

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"proxychan/internal/logging"
	"proxychan/internal/models"
	"strings"
	"time"
)

var cfg = &models.DefaultRuntimeConfig

func TorServiceStart(torSocksAddr string) {
	ctrl, err := getTorController()
	if err != nil {
		logging.GetLogger().Fatalf("Failed to get Tor controller: %v", err)
	}

	// Best-effort service check (informational)
	_ = ctrl.IsTorRunning()

	// Authoritative check
	if torReachable(torSocksAddr) {
		logging.GetLogger().Info("Tor SOCKS is already reachable.")
		return
	}

	fmt.Println("Tor SOCKS is not reachable.")
	fmt.Print("Start Tor service now? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(strings.ToLower(ans))

	if ans != "y" && ans != "yes" {
		fmt.Println("Tor is required. Exiting.")
		os.Exit(1)
	}

	if err := ctrl.StartTor(); err != nil {
		logging.GetLogger().Fatalf("Failed to start Tor: %v", err)
	} else {
		cfg.DisableTorOnExit = true
		logging.GetLogger().Info("Tor service started successfully.")
	}

	// One final check
	ready := false
	for i := 0; i < 3; i++ {
		if torReachable(torSocksAddr) {
			ready = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !ready {
		logging.GetLogger().Fatalf("Tor started but SOCKS is still unreachable.")
	}

}

func torReachable(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

func TorServiceStop() {
	if models.DefaultRuntimeConfig.DisableTorOnExit {
		if ctrl, err := getTorController(); err == nil {
			if err := ctrl.StopTor(); err != nil {
				logging.GetLogger().Errorf("Failed to stop Tor service: %v", err)
			} else {
				logging.GetLogger().Info("Tor service stopped.")
			}
		} else {
			logging.GetLogger().Errorf("Failed to get Tor controller: %v", err)
		}
	}
}
