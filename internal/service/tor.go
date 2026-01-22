package service

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"proxychan/internal/models"
	"strings"
	"time"
)

var cfg = &models.DefaultRuntimeConfig

func TorServiceStart(torSocksAddr string) {
	ctrl, err := getTorController()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Best-effort service check (informational)
	_ = ctrl.IsTorRunning()

	// Authoritative check
	if torReachable(torSocksAddr) {
		return
	}

	fmt.Println("Tor SOCKS is not reachable.")
	fmt.Print("Start Tor service now? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(strings.ToLower(ans))

	if ans != "y" && ans != "yes" {
		fmt.Println("Tor required. Exiting.")
		os.Exit(1)
	}

	if err := ctrl.StartTor(); err != nil {
		fmt.Printf("Failed to start Tor: %v\n", err)
		os.Exit(1)
	} else {
		cfg.DisableTorOnExit = true
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
		fmt.Println("Tor started but SOCKS is still unreachable.")
		os.Exit(1)
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
				fmt.Printf("\nFailed to stop Tor service: %v\n", err)
			} else {
				fmt.Println("\nTor service stopped")
			}
		}
	}
}
