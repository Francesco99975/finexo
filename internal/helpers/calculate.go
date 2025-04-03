package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MonthCalcResults stores monthly calculations
type MonthCalcResults struct {
	MonthName                  string `json:"monthName"`
	ShareAmount                string `json:"shareAmount"`
	Contributions              string `json:"contributions"`
	MonthlyGainedFromPriceInc  string `json:"monthlyGainedFromPriceIncrease"`
	MonthlyGainedFromDividends string `json:"monthlyGainedFromDividends"`
	MonthlyGain                string `json:"monthlyGain"`
	CumGain                    string `json:"cumGain"`
	Balance                    string `json:"balance"`
	Return                     string `json:"return"`
	DRIP                       string `json:"drip"`
}

// YearCalcResults stores yearly calculations
type YearCalcResults struct {
	YearName       string             `json:"yearName"`
	ShareAmount    string             `json:"shareAmount"`
	TotalYearGains string             `json:"totalYearGains"`
	CumGain        string             `json:"cumGain"`
	YoyGrowth      string             `json:"yoyGrowth"`
	TotalGrowth    string             `json:"totalGrowth"`
	Balance        string             `json:"balance"`
	MonthsResults  []MonthCalcResults `json:"monthsResults"`
}

// CalculationResults stores the final output
type CalculationResults struct {
	Profit             string            `json:"profit"`
	TotalContributions string            `json:"totalContributions"`
	FinalBalance       string            `json:"finalBalance"`
	YearResults        []YearCalcResults `json:"yearResults"`
}

func frequencyToMonths(freq string) int {
	switch freq {
	case "monthly":
		return 1
	case "quarterly":
		return 4
	case "semi-annual":
		return 6
	case "annual":
		return 12
	default:
		return 1 // Default to monthly if unrecognized
	}
}

// Function to calculate investment results
func CalculateInvestment(
	stockPrice, dividendYield, expenseRatio, principal, contribution float64,
	contributionFreqStr, dividendFreqStr string, annualPriceIncreasePercent, annualDividendIncreasePercent float64,
	compoundingYears int, payoutMonth int, currency string,
) (CalculationResults, error) {

	// ✅ Convert string frequency inputs to numeric values
	contributionFrequency := frequencyToMonths(contributionFreqStr)
	dividendFrequency := frequencyToMonths(dividendFreqStr)

	// Convert percentages to decimal
	annualPriceIncrease := annualPriceIncreasePercent / 100
	annualDividendIncrease := annualDividendIncreasePercent / 100
	expenseRate := expenseRatio / 100

	// Determine first month (payout month)
	currentMonth := time.Now().Month()
	startMonth := payoutMonth
	if startMonth == 0 {
		startMonth = int(currentMonth) + 1 // Default to next month if not provided
		if startMonth > 12 {
			startMonth = 1
		}
	}

	currentYear := time.Now().Year()

	// Initial values
	totalContributions := 0.0
	totalCumGain := 0.0
	totalShares := principal / stockPrice
	currentStockPrice := stockPrice
	currentDividendYield := dividendYield / 100
	balance := principal
	yearResults := []YearCalcResults{}

	// Loop through each year
	for year := 1; year <= compoundingYears; year++ {
		yearGains := 0.0
		monthResults := []MonthCalcResults{}

		// Loop through each month, starting from `startMonth`
		for month := startMonth; month <= 12; month++ {
			// Check if this month is a contribution month based on frequency
			contributionThisMonth := 0.0
			if (month-startMonth)%contributionFrequency == 0 {
				contributionThisMonth = contribution
			}
			totalContributions += contributionThisMonth

			// Calculate stock price increase
			prevStockPrice := currentStockPrice
			monthlyPriceIncreaseFactor := 1 + (annualPriceIncrease / 12)
			currentStockPrice *= monthlyPriceIncreaseFactor

			// Calculate dividend yield increase
			// prevDividendYield := currentDividendYield
			monthlyDividendIncreaseFactor := 1 + (annualDividendIncrease / 12)
			currentDividendYield *= monthlyDividendIncreaseFactor

			// Calculate gains from price increase
			monthlyGainedFromPriceIncrease := totalShares * (currentStockPrice - prevStockPrice)

			// Calculate dividend earnings
			monthlyGainedFromDividends := 0.0
			if dividendFrequency > 0 && (month-startMonth)%dividendFrequency == 0 {
				// Use CURRENT stock price instead of previous
				monthlyGainedFromDividends = totalShares * currentStockPrice * (currentDividendYield / float64(12/dividendFrequency))
			}

			// Total monthly gain (after expenses)
			monthlyGain := (monthlyGainedFromPriceIncrease + monthlyGainedFromDividends) * (1 - (expenseRate / 12))
			totalCumGain += monthlyGain
			yearGains += monthlyGain

			// Balance update
			balance += contributionThisMonth + monthlyGain

			// Calculate return percentage
			monthlyReturn := 0.0
			if balance-contributionThisMonth > 0 {
				monthlyReturn = (monthlyGain / (balance - contributionThisMonth)) * 100
			}

			// Calculate DRIP (Dividend Reinvestment)
			newShares := 0.0
			DRIPStatus := "N/A"
			if monthlyGainedFromDividends > 0 {
				newShares = monthlyGainedFromDividends / currentStockPrice
				if newShares >= 1 && newShares < 2 {
					DRIPStatus = "DRIP"
				} else if newShares >= 2 {
					DRIPStatus = "DRIPx" + fmt.Sprintf("%.0f", newShares)
				} else {
					DRIPStatus = "NO DRIP"
				}
			}

			// Update total shares
			totalShares += newShares

			// Store month results if it's a contribution or dividend month
			if contributionThisMonth > 0 || (dividendFrequency > 0 && (month-startMonth)%dividendFrequency == 0) {
				formattedTotalContributions, err := FormatPrice(totalContributions, currency)
				if err != nil {
					return CalculationResults{}, err
				}

				formattedMonthlyGainedFromPriceIncrease, err := FormatPrice(monthlyGainedFromPriceIncrease, currency)
				if err != nil {
					return CalculationResults{}, err
				}
				formattedMonthlyGainedFromDividends, err := FormatPrice(monthlyGainedFromDividends, currency)
				if err != nil {
					return CalculationResults{}, err
				}
				formattedMonthlyGain, err := FormatPrice(monthlyGain, currency)
				if err != nil {
					return CalculationResults{}, err
				}
				formattedTotalCumGain, err := FormatPrice(totalCumGain, currency)
				if err != nil {
					return CalculationResults{}, err
				}
				formattedBalance, err := FormatPrice(balance, currency)
				if err != nil {
					return CalculationResults{}, err
				}

				monthResults = append(monthResults, MonthCalcResults{
					MonthName:                  time.Month(month).String(),
					ShareAmount:                fmt.Sprintf("%.2f", totalShares),
					Contributions:              formattedTotalContributions,
					MonthlyGainedFromPriceInc:  formattedMonthlyGainedFromPriceIncrease,
					MonthlyGainedFromDividends: formattedMonthlyGainedFromDividends,
					MonthlyGain:                formattedMonthlyGain,
					CumGain:                    formattedTotalCumGain,
					Balance:                    formattedBalance,
					Return:                     fmt.Sprintf("%.2f%%", monthlyReturn),
					DRIP:                       DRIPStatus,
				})
			}
		}

		// Yearly metrics
		cumGain := totalCumGain
		yoyGrowth := 0.0
		if year > 1 {
			prevBalance, err := parseFloat(strings.ReplaceAll(yearResults[year-2].Balance[4:], ",", ""))
			if err != nil {
				return CalculationResults{}, err
			}
			yoyGrowth = ((balance - prevBalance) / prevBalance) * 100
		} else {
			prevBalance := principal

			yoyGrowth = ((balance - prevBalance) / prevBalance) * 100
		}

		totalGrowth := ((balance - totalContributions) / totalContributions) * 100

		formattedYearGains, err := FormatPrice(yearGains, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		formattedCumGain, err := FormatPrice(cumGain, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		formattedBalance, err := FormatPrice(balance, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		// Store yearly results
		yearResults = append(yearResults, YearCalcResults{
			YearName:       fmt.Sprintf("Year (%d) - %d", year, currentYear+year-1),
			ShareAmount:    fmt.Sprintf("%.2f", totalShares),
			TotalYearGains: formattedYearGains,
			CumGain:        formattedCumGain,
			YoyGrowth:      fmt.Sprintf("%.2f%%", yoyGrowth),
			TotalGrowth:    fmt.Sprintf("%.2f%%", totalGrowth),
			Balance:        formattedBalance,
			MonthsResults:  monthResults,
		})
	}

	formattedTotalContributions, err := FormatPrice(totalContributions, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	formattedFinalBalance, err := FormatPrice(balance, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	formattedProfit, err := FormatPrice(balance-totalContributions, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	// Return final results
	return CalculationResults{
		Profit:             formattedProfit,
		TotalContributions: formattedTotalContributions,
		FinalBalance:       formattedFinalBalance,
		YearResults:        yearResults,
	}, nil
}

func parseFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
