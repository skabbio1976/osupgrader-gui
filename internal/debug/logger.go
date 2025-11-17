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

// Init initierar debug-loggning
func Init() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		return nil // Redan initierad
	}

	// Hitta binärens katalog
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("kunde inte hitta binärens sökväg: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Skapa loggfil i samma mapp som binären
	logPath = filepath.Join(exeDir, "debuglogg.txt")

	// Öppna/skapa loggfil (append mode)
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("kunde inte skapa loggfil %s: %w", logPath, err)
	}

	isEnabled = true

	// Logga start
	writeLog("=== DEBUG LOGGING STARTED ===")
	writeLog("Log file: %s", logPath)
	writeLog("Timestamp: %s", time.Now().Format("2006-01-02 15:04:05"))
	writeLog("")

	return nil
}

// Close stänger loggfilen
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

// Log skriver ett debug-meddelande
func Log(format string, args ...interface{}) {
	if !isEnabled {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	writeLog(format, args...)
}

// LogError skriver ett felmeddelande
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

// LogFunction skriver funktionsanrop med parametrar
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

// LogSuccess skriver ett success-meddelande
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

// writeLog skriver direkt till loggfilen (utan lock, måste anropas inuti lock)
func writeLog(format string, args ...interface{}) {
	if logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] %s\n", timestamp, msg)

	logFile.WriteString(line)
	logFile.Sync() // Flush direkt för att säkerställa att allt skrivs
}

// GetLogPath returnerar sökvägen till loggfilen
func GetLogPath() string {
	return logPath
}

// IsEnabled returnerar om debug-loggning är aktiverad
func IsEnabled() bool {
	return isEnabled
}
