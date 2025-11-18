package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile   *os.File
	logMutex  sync.Mutex
	logPath   string
	isEnabled bool
)

// Init initializes debug logging
func Init() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		return nil // Already initialized
	}

	// Find binary directory
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find binary path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Create log file in same directory as binary
	logPath = filepath.Join(exeDir, "debuglogg.txt")

	// Open/create log file (append mode)
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("could not create log file %s: %w", logPath, err)
	}

	isEnabled = true

	// Log start
	writeLog("=== DEBUG LOGGING STARTED ===")
	writeLog("Log file: %s", logPath)
	writeLog("Timestamp: %s", time.Now().Format("2006-01-02 15:04:05"))
	writeLog("")

	return nil
}

// Close closes the log file
func Close() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		writeLog("")
		writeLog("=== DEBUG LOGGING ENDED ===")
		writeLog("")
		logFile.Close()
		logFile = nil
	}
	isEnabled = false
}

// Log writes a debug message
func Log(format string, args ...interface{}) {
	if !isEnabled {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	writeLog(format, args...)
}

// LogError writes an error message
func LogError(context string, err error, details ...interface{}) {
	if !isEnabled {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	writeLog("ERROR [%s]: %v", context, err)
	for i := 0; i < len(details); i += 2 {
		if i+1 < len(details) {
			writeLog("  %v: %v", details[i], details[i+1])
		}
	}
	writeLog("")
}

// LogFunction writes function calls with parameters
func LogFunction(funcName string, params ...interface{}) {
	if !isEnabled {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	writeLog("CALL %s", funcName)
	for i := 0; i < len(params); i += 2 {
		if i+1 < len(params) {
			writeLog("  %v: %v", params[i], params[i+1])
		}
	}
}

// LogSuccess writes a success message
func LogSuccess(context string, details ...interface{}) {
	if !isEnabled {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	writeLog("SUCCESS [%s]", context)
	for i := 0; i < len(details); i += 2 {
		if i+1 < len(details) {
			writeLog("  %v: %v", details[i], details[i+1])
		}
	}
	writeLog("")
}

// writeLog writes directly to log file (without lock, must be called inside lock)
func writeLog(format string, args ...interface{}) {
	if logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] %s\n", timestamp, msg)

	logFile.WriteString(line)
	logFile.Sync() // Flush immediately to ensure everything is written
}

// GetLogPath returns the path to the log file
func GetLogPath() string {
	return logPath
}

// IsEnabled returns whether debug logging is enabled
func IsEnabled() bool {
	return isEnabled
}
