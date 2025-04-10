package tools

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Francesco99975/finexo/internal/helpers"
)

func ReadAllSeeds() ([]string, error) {
	// Define the directory containing CSV files
	seedDir := "seeds"
	var seeds []string

	// Read all CSV files in the directory
	err := filepath.Walk(seedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Process only .csv files
		if strings.HasSuffix(strings.ToLower(info.Name()), ".csv") {
			payload, err := readSeed(path)
			if err != nil {
				return err
			}
			seeds = append(seeds, payload...)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Error reading seeds: %v", err)
	}

	filterSeeds := helpers.FilteredSlice(seeds, func(s string) bool {
		return len(s) < 6
	})

	return filterSeeds, nil
}

func readSeed(path string) ([]string, error) {
	// Open the CSV file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %v", err)
	}
	defer file.Close()

	recordsSet := make(map[string]bool)

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
		if err != nil {
			break // EOF
		}
		recordsSet[normalizeSeed(record[columnIndex])] = true
		if strings.Contains(path, "canadian-stocks-us-stocks") {
			recordsSet[normalizeSeed(record[columnIndex])+".TO"] = true
		}
	}

	records := make([]string, 0, len(recordsSet))
	for record := range recordsSet {
		records = append(records, record)
	}

	return records, nil
}

func normalizeSeed(seed string) string {
	seed = strings.ToUpper(seed)
	seed = strings.TrimSpace(seed)
	seed = strings.ReplaceAll(seed, " ", "")

	seed = strings.Replace(seed, ".", "-", 1)
	seed = strings.ReplaceAll(seed, "-TO", ".TO")
	seed = strings.ReplaceAll(seed, "-NE", ".NE")
	seed = strings.ReplaceAll(seed, "-L", ".L")
	seed = strings.ReplaceAll(seed, "-V", ".V")
	seed = strings.ReplaceAll(seed, "-T", ".T")
	seed = strings.ReplaceAll(seed, "-MI", ".MI")
	seed = strings.ReplaceAll(seed, "-F", ".F")
	return seed
}
