package helpers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
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
	SID                string            `json:"sid"`
	Principal          string            `json:"principal"`
	Rate               string            `json:"rate"`
	RateFreq           string            `json:"rateFreq"`
	Currency           string            `json:"currency"`
	Profit             string            `json:"profit"`
	TotalContributions string            `json:"totalContributions"`
	ContribFreq        string            `json:"contribFreq"`
	FinalBalance       string            `json:"finalBalance"`
	YearResults        []YearCalcResults `json:"yearResults"`
}

func DecodeResults(encoded string) (CalculationResults, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return CalculationResults{}, err
	}

	gzipped, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return CalculationResults{}, err
	}

	gzipped.Close()

	var results CalculationResults
	if err := json.NewDecoder(gzipped).Decode(&results); err != nil {
		return CalculationResults{}, err
	}

	return results, nil
}

func (cr CalculationResults) Encoded() (string, error) {
	jsonBytes, err := json.Marshal(cr)
	if err != nil {
		return "", err
	}

	var gzipped bytes.Buffer
	gz := gzip.NewWriter(&gzipped)
	if _, err := gz.Write(jsonBytes); err != nil {
		return "", err
	}
	gz.Close()

	encoded := base64.StdEncoding.EncodeToString(gzipped.Bytes())

	return encoded, nil
}

func frequencyToMonths(freq string) int {
	switch freq {
	case "monthly":
		return 1
	case "quarterly":
		return 3
	case "semi-annual":
		return 6
	case "annual":
		return 12
	default:
		return 1 // Default to monthly if unrecognized
	}
}

func compoundingPeriodsPerYear(freq string) int {
	switch freq {
	case "daily":
		return 365
	case "monthly":
		return 12
	case "quarterly":
		return 4
	case "semi-annual":
		return 2
	case "annual":
		return 1
	default:
		return 12 // Default to monthly if unrecognized
	}
}

func CalculateHISAInvestment(principal, contribution float64, contributionFreqStr, compoundingFreqStr string, annualInterestRate float64, compoundingYears int, currency string) (CalculationResults, error) {
	// Convert frequency strings to integers
	contributionFreq := frequencyToMonths(contributionFreqStr)
	n := float64(compoundingPeriodsPerYear(compoundingFreqStr))

	monthlyInterestfactor := math.Pow(1+(annualInterestRate/100)/n, n/12)

	balance := principal
	totalContributions := principal
	cumGain := 0.0

	currentYear := time.Now().Year()

	yearResults := []YearCalcResults{}

	// Loop through each year
	for year := 1; year <= compoundingYears; year++ {
		monthResults := []MonthCalcResults{}
		var startMonth int
		if year == 1 {
			// For the first year, start from the current month
			startMonth = int(time.Now().Month())
		} else {
			// For subsequent years, start from January
			startMonth = 1
		}

		// Loop through each month, starting from `startMonth`
		for month := startMonth; month <= 12; month++ {

			monthName := time.Month(month).String()

			balanceBeginning := balance

			contributionThisMonth := 0.0
			if (month-1)%contributionFreq == 0 {
				contributionThisMonth = contribution
				balance += contributionThisMonth
				totalContributions += contributionThisMonth
			}

			interestEarned := balanceBeginning * (monthlyInterestfactor - 1)
			balance *= monthlyInterestfactor
			cumGain += interestEarned

			monthlyGain := interestEarned

			returnPercent := 0.0
			if balanceBeginning > 0 {
				returnPercent = (balance - balanceBeginning) / balanceBeginning * 100
			}

			formattedTotalContributions, err := FormatPrice(totalContributions, currency)
			if err != nil {
				return CalculationResults{}, err
			}

			formattedMonthlyGainedFromPriceIncrease, err := FormatPrice(interestEarned, currency)
			if err != nil {
				return CalculationResults{}, err
			}

			formattedMonthlyGain, err := FormatPrice(monthlyGain, currency)
			if err != nil {
				return CalculationResults{}, err
			}
			formattedTotalCumGain, err := FormatPrice(cumGain, currency)
			if err != nil {
				return CalculationResults{}, err
			}
			formattedBalance, err := FormatPrice(balance, currency)
			if err != nil {
				return CalculationResults{}, err
			}

			monthResults = append(monthResults, MonthCalcResults{
				MonthName:                  monthName,
				ShareAmount:                "N/A",
				Contributions:              formattedTotalContributions,
				MonthlyGainedFromPriceInc:  formattedMonthlyGainedFromPriceIncrease,
				MonthlyGainedFromDividends: "N/A",
				MonthlyGain:                formattedMonthlyGain,
				CumGain:                    formattedTotalCumGain,
				Balance:                    formattedBalance,
				Return:                     fmt.Sprintf("%.2f%%", returnPercent),
				DRIP:                       "N/A",
			})

		}

		totalYearGains := 0.0
		for _, monthResult := range monthResults {
			mg, err := parseFloat(strings.ReplaceAll(monthResult.MonthlyGain[4:], ",", ""))
			if err != nil {
				return CalculationResults{}, err
			}
			totalYearGains += mg
		}

		currentYearBalance := balance

		yoyGrowth := 0.0

		if year > 1 {
			prevBalance, err := parseFloat(strings.ReplaceAll(yearResults[year-2].Balance[4:], ",", ""))
			if err != nil {
				return CalculationResults{}, err
			}
			yoyGrowth = ((currentYearBalance - prevBalance) / prevBalance) * 100
		} else {
			prevBalance := totalContributions

			yoyGrowth = ((currentYearBalance - prevBalance) / prevBalance) * 100
		}

		totalGrowth := (currentYearBalance - totalContributions) / totalContributions * 100

		formattedYearGains, err := FormatPrice(totalYearGains, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		formattedCumGain, err := FormatPrice(cumGain, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		formattedBalance, err := FormatPrice(currentYearBalance, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		// Store yearly results
		yearResults = append(yearResults, YearCalcResults{
			YearName:       fmt.Sprintf("Year (%d) - %d", year, currentYear+year-1),
			ShareAmount:    "N/A",
			TotalYearGains: formattedYearGains,
			CumGain:        formattedCumGain,
			YoyGrowth:      fmt.Sprintf("%.2f%%", yoyGrowth),
			TotalGrowth:    fmt.Sprintf("%.2f%%", totalGrowth),
			Balance:        formattedBalance,
			MonthsResults:  monthResults,
		})

	}

	finalBalance := balance

	formattedTotalContributions, err := FormatPrice(totalContributions, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	formattedFinalBalance, err := FormatPrice(finalBalance, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	formattedProfit, err := FormatPrice(finalBalance-totalContributions, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	formattedPrincial, err := FormatPrice(principal, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	// Return final results
	return CalculationResults{
		Principal:          formattedPrincial,
		Rate:               fmt.Sprintf("%.2f%%", annualInterestRate),
		RateFreq:           compoundingFreqStr,
		Currency:           currency,
		Profit:             formattedProfit,
		TotalContributions: formattedTotalContributions,
		ContribFreq:        contributionFreqStr,
		FinalBalance:       formattedFinalBalance,
		YearResults:        yearResults,
	}, nil
}

// Function to calculate investment results
func CalculateInvestment(
	stockPrice, dividendYield, expenseRatio, principal, contribution float64,
	contributionFreqStr, dividendFreqStr string, annualPriceIncreasePercent, annualDividendIncreasePercent float64,
	compoundingYears int, payoutMonth int, currency string,
) (CalculationResults, error) {

	// Log input parameters
	log.Debugf("stockPrice: %f", stockPrice)
	log.Debugf("dividendYield: %f", dividendYield)
	log.Debugf("expenseRatio: %f", expenseRatio)
	log.Debugf("principal: %f", principal)
	log.Debugf("contribution: %f", contribution)
	log.Debugf("contributionFreqStr: %s", contributionFreqStr)
	log.Debugf("dividendFreqStr: %s", dividendFreqStr)
	log.Debugf("annualPriceIncreasePercent: %f", annualPriceIncreasePercent)
	log.Debugf("annualDividendIncreasePercent: %f", annualDividendIncreasePercent)
	log.Debugf("compoundingYears: %d", compoundingYears)
	log.Debugf("payoutMonth: %d", payoutMonth)
	log.Debugf("currency: %s", currency)

	// ✅ Convert string frequency inputs to numeric values
	contributionFrequency := frequencyToMonths(contributionFreqStr)
	dividendFrequency := frequencyToMonths(dividendFreqStr)

	// Calculate number of payments per year
	mDiv := 0
	if dividendYield > 0 {
		mDiv = 12 / dividendFrequency
	}

	// Convert percentages to decimal
	annualPriceIncrease := annualPriceIncreasePercent / 100
	log.Debugf("annualPriceIncrease: %f", annualPriceIncrease)
	annualDividendIncrease := annualDividendIncreasePercent / 100
	log.Debugf("annualDividendIncrease: %f", annualDividendIncrease)

	// Initial values
	totalContributions := principal
	cumGain := 0.0
	shares := principal / stockPrice
	monthlyPriceGrowth := math.Pow(1+annualPriceIncrease, 1.0/12.0)
	log.Debugf("monthlyPriceGrowth: %f", monthlyPriceGrowth)
	D0 := ((dividendYield - expenseRatio) / 100) * stockPrice
	log.Debugf("D0: %f", D0)
	totalMonth := 0

	currentYear := time.Now().Year()

	yearResults := []YearCalcResults{}

	// Loop through each year
	for year := 1; year <= compoundingYears; year++ {
		// Annual Dividend per share this year
		Dy := D0 * math.Pow(1+annualDividendIncrease, float64(year-1))
		log.Debugf("Year %d Dy: %f", year, Dy)
		dividendPerPeriod := Dy / float64(mDiv)
		log.Debugf("Year %d dividendPerPeriod: %f", year, dividendPerPeriod)

		monthResults := []MonthCalcResults{}

		var startMonth int
		if year == 1 {
			// Determine first month (payout month)
			currentMonth := time.Now().Month()
			startMonth = payoutMonth
			if startMonth == 0 {
				startMonth = int(currentMonth) + 1 // Default to next month if not provided
				if startMonth > 12 {
					startMonth = 1
				}
			}
		} else {
			// If not the first year, start from January
			startMonth = 1
		}

		// Loop through each month, starting from `startMonth`
		for month := startMonth; month <= 12; month++ {
			totalMonth++
			monthName := time.Month(month).String()

			stockPriceBegin := stockPrice
			sharesBefore := shares
			balanceBeginning := sharesBefore * stockPriceBegin
			// Check if this month is a contribution month based on frequency
			contributionThisMonth := 0.0
			if (totalMonth-1)%contributionFrequency == 0 {
				contributionThisMonth = contribution
				sharesBought := contribution / stockPrice
				shares += sharesBought
				totalContributions += contributionThisMonth
			}

			dividendReceived := 0.0
			sharesBoughtDividend := 0.0
			if dividendYield > 0 && (totalMonth-1)%dividendFrequency == 0 {
				dividendReceived = shares * dividendPerPeriod
				log.Debugf("Year %d Month %s dividendReceived: %f", year, monthName, dividendReceived)
				sharesBoughtDividend = dividendReceived / stockPriceBegin
				log.Debugf("Year %d Month %s sharesBoughtDividend: %f", year, monthName, sharesBoughtDividend)
				shares += sharesBoughtDividend
				log.Debugf("Year %d Month %s shares: %f", year, monthName, shares)
				cumGain += dividendReceived
			}

			balanceBeforePriceChange := shares * stockPriceBegin
			log.Debugf("Year %d Month %s balanceBeforePriceChange: %f", year, monthName, balanceBeforePriceChange)
			stockPrice *= monthlyPriceGrowth
			log.Debugf("Year %d Month %s stockPrice: %f", year, monthName, stockPrice)
			balanceEnd := shares * stockPrice
			log.Debugf("Year %d Month %s balanceEnd: %f", year, monthName, balanceEnd)

			//Calculate Gains
			monthlyGainsFromDividends := dividendReceived
			monthlyGainsFromPriceIncrease := balanceEnd - balanceBeforePriceChange
			log.Debugf("Year %d Month %s monthlyGainsFromDividends: %f", year, monthName, monthlyGainsFromDividends)
			monthlyGains := monthlyGainsFromDividends + monthlyGainsFromPriceIncrease
			log.Debugf("Year %d Month %s monthlyGains: %f", year, monthName, monthlyGains)
			cumGain += monthlyGainsFromPriceIncrease
			log.Debugf("Year %d Month %s cumGain: %f", year, monthName, cumGain)

			// Calculate return percentage
			returnPercent := 0.0
			if balanceBeginning > 0 {
				returnPercent = (balanceEnd - balanceBeginning) / balanceBeginning * 100
			}

			// Calculate DRIP (Dividend Reinvestment)

			DRIPStatus := "N/A"

			if dividendYield > 0 {
				if sharesBoughtDividend >= 1 && sharesBoughtDividend < 2 {
					DRIPStatus = "DRIP"
				} else if sharesBoughtDividend >= 2 {
					DRIPStatus = "DRIPx" + fmt.Sprintf("%.0f", sharesBoughtDividend)
				} else {
					DRIPStatus = "NO DRIP"
				}
			}

			// Store month results if it's a contribution or dividend month

			formattedTotalContributions, err := FormatPrice(totalContributions, currency)
			if err != nil {
				return CalculationResults{}, err
			}

			formattedMonthlyGainedFromPriceIncrease, err := FormatPrice(monthlyGainsFromPriceIncrease, currency)
			if err != nil {
				return CalculationResults{}, err
			}
			formattedMonthlyGainedFromDividends, err := FormatPrice(monthlyGainsFromDividends, currency)
			if err != nil {
				return CalculationResults{}, err
			}
			formattedMonthlyGain, err := FormatPrice(monthlyGains, currency)
			if err != nil {
				return CalculationResults{}, err
			}
			formattedTotalCumGain, err := FormatPrice(cumGain, currency)
			if err != nil {
				return CalculationResults{}, err
			}
			formattedBalance, err := FormatPrice(balanceEnd, currency)
			if err != nil {
				return CalculationResults{}, err
			}

			monthResults = append(monthResults, MonthCalcResults{
				MonthName:                  monthName,
				ShareAmount:                fmt.Sprintf("%.2f", shares),
				Contributions:              formattedTotalContributions,
				MonthlyGainedFromPriceInc:  formattedMonthlyGainedFromPriceIncrease,
				MonthlyGainedFromDividends: formattedMonthlyGainedFromDividends,
				MonthlyGain:                formattedMonthlyGain,
				CumGain:                    formattedTotalCumGain,
				Balance:                    formattedBalance,
				Return:                     fmt.Sprintf("%.2f%%", returnPercent),
				DRIP:                       DRIPStatus,
			})

		}

		// Yearly metrics
		totalYearGains := 0.0

		for _, monthResult := range monthResults {
			mg, err := parseFloat(strings.ReplaceAll(monthResult.MonthlyGain[4:], ",", ""))
			if err != nil {
				return CalculationResults{}, err
			}
			totalYearGains += mg
		}

		currentYearBalance := shares * stockPrice

		yoyGrowth := 0.0
		if year > 1 {
			prevBalance, err := parseFloat(strings.ReplaceAll(yearResults[year-2].Balance[4:], ",", ""))
			if err != nil {
				return CalculationResults{}, err
			}
			yoyGrowth = ((currentYearBalance - prevBalance) / prevBalance) * 100
		} else {
			prevBalance := totalContributions

			yoyGrowth = ((currentYearBalance - prevBalance) / prevBalance) * 100
		}

		totalGrowth := (currentYearBalance - totalContributions) / totalContributions * 100

		formattedYearGains, err := FormatPrice(totalYearGains, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		formattedCumGain, err := FormatPrice(cumGain, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		formattedBalance, err := FormatPrice(currentYearBalance, currency)
		if err != nil {
			return CalculationResults{}, err
		}

		// Store yearly results
		yearResults = append(yearResults, YearCalcResults{
			YearName:       fmt.Sprintf("Year (%d) - %d", year, currentYear+year-1),
			ShareAmount:    fmt.Sprintf("%.2f", shares),
			TotalYearGains: formattedYearGains,
			CumGain:        formattedCumGain,
			YoyGrowth:      fmt.Sprintf("%.2f%%", yoyGrowth),
			TotalGrowth:    fmt.Sprintf("%.2f%%", totalGrowth),
			Balance:        formattedBalance,
			MonthsResults:  monthResults,
		})
	}

	formattedPrincial, err := FormatPrice(principal, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	formattedTotalContributions, err := FormatPrice(totalContributions, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	finalBalance := shares * stockPrice

	formattedFinalBalance, err := FormatPrice(finalBalance, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	formattedProfit, err := FormatPrice(finalBalance-totalContributions, currency)
	if err != nil {
		return CalculationResults{}, err
	}

	// Return final results
	return CalculationResults{
		Principal:          formattedPrincial,
		Rate:               fmt.Sprintf("%.2f%%", dividendYield-expenseRatio),
		RateFreq:           dividendFreqStr,
		Currency:           currency,
		Profit:             formattedProfit,
		TotalContributions: formattedTotalContributions,
		ContribFreq:        contributionFreqStr,
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
