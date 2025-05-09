package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/labstack/gommon/log"
)

// Structured Severity "enum"
type Severity struct {
	PANIC SeverityType
	ERROR SeverityType
	WARN  SeverityType
	INFO  SeverityType
	DEBUG SeverityType
}

// SeverityType defines the type for severity levels
type SeverityType string

// Instantiate Severity constants
var SeverityLevels = &Severity{
	PANIC: "PANIC",
	ERROR: "ERROR",
	WARN:  "WARN",
	INFO:  "INFO",
	DEBUG: "DEBUG",
}

type Reporter struct {
	lock     sync.Mutex
	filePath string
	file     *os.File
}

func NewReporter(filePath string) (*Reporter, error) {
	// Create directory structure if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		filePath: filePath,
		file:     file,
	}, nil
}

// Report writes a log entry to the file
func (r *Reporter) Report(level SeverityType, message string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("%s [%s] %s\n", timestamp, level, message)

	_, err := r.file.WriteString(entry)
	if err != nil {
		return err
	}

	return nil
}

// Close the report file
func (r *Reporter) Close() error {
	return r.file.Close()
}

func Cleanup(frequency time.Duration) {
	// Remove old report files in the report directory
	syncFiles, err := os.ReadDir("reports/sync")
	if err != nil {
		log.Errorf("Failed to read report directory: %v", err)
		return
	}
	for _, file := range syncFiles {
		if file.IsDir() {
			continue
		}

		fileInfo, err := file.Info()
		if err != nil {
			log.Errorf("Failed to get file info: %v", err)
			continue
		}
		if time.Since(fileInfo.ModTime()) > frequency {
			err := os.Remove(filepath.Join("reports/sync", file.Name()))
			if err != nil {
				log.Errorf("Failed to remove old report file: %v", err)
			}
		}
	}

	discoveryFiles, err := os.ReadDir("reports/discovery")
	if err != nil {
		log.Errorf("Failed to read report directory: %v", err)
		return
	}
	for _, file := range discoveryFiles {
		if file.IsDir() {
			continue
		}

		fileInfo, err := file.Info()
		if err != nil {
			log.Errorf("Failed to get file info: %v", err)
			continue
		}
		if time.Since(fileInfo.ModTime()) > frequency {
			err := os.Remove(filepath.Join("reports/discovery", file.Name()))
			if err != nil {
				log.Errorf("Failed to remove old report file: %v", err)
			}
		}
	}
}
