package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

const logDir = "/var/log/proxychan"

// SetupLogger initializes logrus with log rotation and date-based log file naming
func SetupLogger() {

	if err := os.MkdirAll(logDir, 0750); err != nil {
		fmt.Println("Error creating logs directory:", err)
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
