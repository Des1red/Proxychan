package cmd

import (
	"fmt"
	"os"
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
