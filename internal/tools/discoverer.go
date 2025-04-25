package tools

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/labstack/gommon/log"
)

type Discoverer struct {
	lock        sync.Mutex
	file        *os.File
	memory      map[string]bool
	discoveries uint
}

func NewDiscoverer() (*Discoverer, error) {
	file, err := os.OpenFile("seeds/discovered.csv", os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	mem := make(map[string]bool)

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("Error reading header row: %v", err)
	}

	// Find the index of the column containing "symbol" or "ticker"
	var columnIndex int = -1
	for i, header := range headers {
		lowerHeader := strings.ToLower(header) // Convert to lowercase for case insensitivity
		if strings.Contains(lowerHeader, "symbol") || strings.Contains(lowerHeader, "ticker") {
			columnIndex = i
			break
		}
	}

	// If no matching column is found, exit
	if columnIndex == -1 {
		return nil, fmt.Errorf("No column containing 'symbol' or 'ticker' found")
	}

	// Read and extract values from the found column
	for {
		record, err := reader.Read()
		if err == io.EOF { // Fix EOF handling
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error reading CSV: %v", err)
		}

		normalized := normalizeSeed(record[columnIndex]) // Normalize before storing
		mem[normalized] = true
	}

	file.Close()

	fileInstance, err := os.OpenFile("seeds/discovered.csv", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	log.Infof("Loaded %d seeds", len(mem))

	return &Discoverer{
		file:        fileInstance,
		memory:      mem,
		discoveries: 0,
	}, nil
}

// Report writes a log entry to the file
func (r *Discoverer) Collect(seed string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	// Normalize the seed
	seed = normalizeSeed(seed)

	if _, exists := r.memory[seed]; exists {
		return nil
	}
	r.memory[seed] = true
	r.discoveries++
	log.Infof("Discovered new SEED >>>: %s", seed)

	writer := csv.NewWriter(r.file)
	defer writer.Flush()

	if err := writer.Write([]string{seed}); err != nil {
		return err
	}

	return nil
}

// Close the report file
func (r *Discoverer) Close() error {
	log.Infof("Discovered %d new seeds", r.discoveries)
	return r.file.Close()
}
