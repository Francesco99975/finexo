package helpers

import (
	"fmt"
	"os"
	"sync"
	"time"
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
