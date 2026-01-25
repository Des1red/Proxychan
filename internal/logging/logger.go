package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func LogDir() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return "/var/log/proxychan", nil
	case "darwin":
		return "/Library/Logs/ProxyChan", nil
	case "windows":
		pd := os.Getenv("ProgramData")
		if pd == "" {
			return "", fmt.Errorf("ProgramData not set")
		}
		return filepath.Join(pd, "ProxyChan", "logs"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func fallbackLogger() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stderr)
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetLevel(logrus.WarnLevel)
}

// SetupLogger initializes logrus with log rotation and date-based log file naming
func SetupLogger() {

	logDir, err := LogDir()
	if err != nil {
		fmt.Println("log path error:", err)
		fallbackLogger()
		return
	}

	if err := os.MkdirAll(logDir, 0750); err != nil {
		fmt.Println("Error creating logs directory:", err)
		fallbackLogger()
		return
	}

	// Get the current date for log file naming
	currentDate := time.Now().Format("2006-01-02") // Format: YYYY-MM-DD
	logFilePath := filepath.Join(logDir, currentDate+".log")

	// Set up the logger with log rotation
	Logger = logrus.New()

	Logger.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath, // Log file path with date
		MaxSize:    10,          // Max size (in MB) before rotating
		MaxBackups: 3,           // Keep up to 3 backups
		MaxAge:     30,          // Retain backups for 28 days
	})

	// Set the log formatter to JSONFormatter for structured logs
	Logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true, // Add pretty print (optional)
	})

	// Optionally set the log level (you can choose debug, info, warn, etc.)
	Logger.SetLevel(logrus.InfoLevel)
}

// GetLogger returns the global logrus Logger instance
func GetLogger() *logrus.Logger {
	if Logger == nil {
		fallbackLogger()
	}
	return Logger
}

// LogConnection logs details of each connection attempt
func LogConnection(username, clientIP, bindAddress, targetAddress, status string) {
	Logger.WithFields(logrus.Fields{
		"user":       username,
		"client_ip":  clientIP,
		"proxy_bind": bindAddress,
		"target":     targetAddress,
		"status":     status,
	}).Info("Proxy connection")
}

// LogAuthFailure logs failed authentication attempts
func LogAuthFailure(clientIP, bindAddress string) {
	Logger.WithFields(logrus.Fields{
		"client_ip":  clientIP,
		"proxy_bind": bindAddress,
	}).Warn("Failed authentication attempt")
}
