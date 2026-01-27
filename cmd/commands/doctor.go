package commands

import (
	"fmt"
	"os"
	"proxychan/internal/server"
	"runtime"
)

func runDoctor(dbPath, logPath string) {
	fmt.Println("ProxyChan Doctor Report")
	fmt.Println("-----------------------")
	fmt.Printf("OS                : %s\n", runtime.GOOS)
	fmt.Printf("Running as        : %s (uid=%d)\n", os.Getenv("USER"), os.Getuid())

	fmt.Println("\nDatabase")
	checkPath(dbPath)

	fmt.Println("\nLogs")
	checkPath(logPath)

	fmt.Println("\nRuntime")
	checkRuntime()
}

func checkPath(path string) {
	_, err := os.Stat(path)
	if err != nil {
		fmt.Printf("  Path            : %s\n", path)
		fmt.Printf("  Exists          : no (%v)\n", err)
		return
	}

	fmt.Printf("  Path            : %s\n", path)
	fmt.Printf("  Exists          : yes\n")

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		fmt.Printf("  Writable        : no (%v)\n", err)
		return
	}
	f.Close()
	fmt.Printf("  Writable        : yes\n")
}

func checkRuntime() {
	fmt.Println("\nRuntime")

	count, err := server.GetActiveConnectionCount()
	if err != nil {
		fmt.Println("  Admin endpoint  : unreachable")
		fmt.Printf("  Error           : %v\n", err)
		return
	}

	fmt.Println("  Admin path	: localhost:6060")
	fmt.Println("  Admin endpoint  : reachable")
	fmt.Printf("  Active tunnels  : %d\n", count)
}
