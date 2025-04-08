package tools

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/Francesco99975/finexo/internal/helpers"
)

// GenerateCSV generates a CSV file with a PDF-like layout from CalculationResults
func GenerateCSV(results helpers.CalculationResults) (string, error) {
	// Generate filename with SID and timestamp
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s-%d.csv", results.SID, timestamp)

	// Open the file for writing
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Initialize CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write Overall Results Section
	overallData := [][]string{
		{"SID", results.SID},
		{"Principal", results.Principal},
		{"Rate", results.Rate},
		{"Rate Frequency", results.RateFreq},
		{"Currency", results.Currency},
		{"Profit", results.Profit},
		{"Total Contributions", results.TotalContributions},
		{"Contribution Frequency", results.ContribFreq},
		{"Final Balance", results.FinalBalance},
	}
	for _, row := range overallData {
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write overall data: %v", err)
		}
	}

	// Add blank row after overall results
	if err := writer.Write([]string{""}); err != nil {
		return "", fmt.Errorf("failed to write blank row: %v", err)
	}

	// Write Yearly and Monthly Breakdowns
	for i, year := range results.YearResults {
		// Write Yearly Summary
		yearData := [][]string{
			{"Year", year.YearName},
			{"Share Amount", year.ShareAmount},
			{"Total Year Gains", year.TotalYearGains},
			{"Cumulative Gain", year.CumGain},
			{"YoY Growth", year.YoyGrowth},
			{"Total Growth", year.TotalGrowth},
			{"Balance", year.Balance},
		}
		for _, row := range yearData {
			if err := writer.Write(row); err != nil {
				return "", fmt.Errorf("failed to write year data: %v", err)
			}
		}

		// Add blank row after yearly summary
		if err := writer.Write([]string{""}); err != nil {
			return "", fmt.Errorf("failed to write blank row: %v", err)
		}

		// Write Monthly Table Header
		monthHeaders := []string{
			"Month", "Shares", "Contributions", "Price Gain", "Dividends Gain",
			"Monthly Gain", "Cumulative Gain", "Balance", "Return", "DRIP",
		}
		if err := writer.Write(monthHeaders); err != nil {
			return "", fmt.Errorf("failed to write month headers: %v", err)
		}

		// Write Monthly Data
		for _, month := range year.MonthsResults {
			monthRow := []string{
				month.MonthName,
				month.ShareAmount,
				month.Contributions,
				month.MonthlyGainedFromPriceInc,
				month.MonthlyGainedFromDividends,
				month.MonthlyGain,
				month.CumGain,
				month.Balance,
				month.Return,
				month.DRIP,
			}
			if err := writer.Write(monthRow); err != nil {
				return "", fmt.Errorf("failed to write month data: %v", err)
			}
		}

		// Add blank row after monthly table (unless it’s the last year)
		if i != len(results.YearResults)-1 {
			if err := writer.Write([]string{""}); err != nil {
				return "", fmt.Errorf("failed to write blank row: %v", err)
			}
		}
	}

	return filename, nil
}
