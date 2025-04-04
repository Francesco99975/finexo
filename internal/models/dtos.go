package models

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/jmoiron/sqlx"
)

type SecuritySearchView struct {
	Ticker   string `json:"ticker"`
	Exchange string `json:"exchange"`
	Title    string `json:"title"` // derived from fullname
	Price    string `json:"price"`
	Typology string `json:"typology"`
	Currency string `json:"currency"`
}

func (s *SecuritySearchView) Scan(rows *sqlx.Rows) error {

	var price int

	// Scan all fields from the row
	err := rows.Scan(
		&s.Ticker, &s.Exchange, &s.Title,
		&price, &s.Typology, &s.Currency,
	)
	if err != nil {
		return err
	}

	// Format price
	priceStr, err := helpers.FormatPrice(float64(price)/100.0, s.Currency)
	if err != nil {
		return err
	}

	s.Price = priceStr

	return nil
}

type SelectedSecurityView struct {
	Ticker                 string `json:"ticker"`
	Exchange               string `json:"exchange"`
	Fullname               string `json:"fullname"`
	Price                  string `json:"price"`
	Typology               string `json:"typology"`
	Currency               string `json:"currency"`
	Target                 string `json:"target"`
	Yield                  string `json:"yield"`
	AnnualPayout           string `json:"annualPayout"`
	PayoutRatio            string `json:"payoutRatio"`
	Frequency              string `json:"frequency"`
	Family                 string `json:"family"`
	ExpenseRatio           string `json:"expenseRatio"`
	ProjectedPriceIncrease string `json:"projectedPriceIncrease"`
	ProjectedYieldIncrease string `json:"projectedYieldIncrease"`
}

func (s *SelectedSecurityView) Scan(rows *sqlx.Rows) error {

	var price int
	var target NullableInt
	var yield int
	var annualPayout NullableInt
	var payoutRatio NullableInt
	var er NullableInt

	// Scan all fields from the row
	err := rows.Scan(
		&s.Ticker, &s.Exchange, &s.Fullname,
		&price, &s.Typology, &s.Currency, &target,
		&yield, &annualPayout, &payoutRatio, &s.Frequency, &s.Family, &er,
	)
	if err != nil {
		return err
	}

	if s.Frequency == "unknown" {
		s.Frequency = "Monthly"
	} else {
		s.Frequency = helpers.Capitalize(s.Frequency)
	}

	// Format price
	priceStr, err := helpers.FormatPrice(float64(price)/100.0, s.Currency)
	if err != nil {
		return err
	}
	s.Price = priceStr

	// Format yield
	s.Yield = fmt.Sprintf("%.2f%%", float64(yield)/100.0)

	if er.Valid {
		// Format expense ratio
		s.ExpenseRatio = fmt.Sprintf("%.2f%%", float64(er.Int64)/100.0)
	} else {
		s.ExpenseRatio = "N/A"
	}

	defaultProjectedYears := 10.0

	if target.Valid {
		targetStr, err := helpers.FormatPrice(float64(target.Int64)/100.0, s.Currency)
		if err != nil {
			return err
		}

		s.Target = targetStr

		if s.Typology != "ETF" {
			s.ProjectedPriceIncrease = fmt.Sprintf("%.2f", (math.Pow(float64(target.Int64/100)/float64(price/100), 1.0/defaultProjectedYears)-1.0)*100.0)
		}

	} else {
		s.Target = "N/A"
		s.ProjectedPriceIncrease = "0"
	}

	if annualPayout.Valid && payoutRatio.Valid {

		annualPayoutStr, err := helpers.FormatPrice(float64(annualPayout.Int64)/100.0, s.Currency)
		if err != nil {
			return err
		}

		s.AnnualPayout = annualPayoutStr

		s.PayoutRatio = fmt.Sprintf("%.2f%%", float64(payoutRatio.Int64)/100.0)

		apf := float64(annualPayout.Int64) / 100.0
		prf := float64(payoutRatio.Int64) / 100.0 / 100.0
		pricef := float64(price) / 100.0

		eps := apf / prf
		roe := eps / pricef

		s.ProjectedYieldIncrease = fmt.Sprintf("%.2f", 100.0*(roe*(1-prf)))
	} else {
		s.AnnualPayout = "N/A"
		s.PayoutRatio = "N/A"
		s.ProjectedYieldIncrease = "0"
	}

	return nil
}

type CalcInput struct {
	SID              string  `form:"sid"`
	Rate             float64 `form:"rate"`
	Principal        float64 `form:"principal"`
	ContribFrequency string  `form:"contribfrequency"`
	Contribution     float64 `form:"contribution"`
	ExReturn         float64 `form:"exreturn"`
	PriceMod         float64 `form:"pricemod"`
	YieldMod         float64 `form:"yieldmod"`
	Years            int     `form:"years"`
}

type SecurityVars struct {
	Price        float64
	Currency     string
	Yield        float64
	Frequency    string
	ExpenseRatio float64
	PayoutMonth  int
}

func (s *SecurityVars) Scan(rows *sqlx.Rows) error {

	var price int
	var yield int
	var er NullableInt
	var payoutDate NullableTime

	// Scan all fields from the row
	err := rows.Scan(
		&price, &s.Currency,
		&yield, &s.Frequency, &er, &payoutDate,
	)
	if err != nil {
		return err
	}

	if s.Frequency == "unknown" {
		s.Frequency = "monthly"
	}

	// Format price
	s.Price = float64(price) / 100.0

	// Format yield
	s.Yield = float64(yield) / 100.0

	if er.Valid {
		// Format expense ratio
		s.ExpenseRatio = float64(er.Int64) / 100
	} else {
		s.ExpenseRatio = 0
	}

	if payoutDate.Valid {
		s.PayoutMonth = int(payoutDate.Time.Month())
	} else {
		s.PayoutMonth = 0
	}

	return nil
}

type SecParams struct {
	Exchange        []string  `query:"exchange"`
	Country         []string  `query:"country"`
	Currency        []string  `query:"currency"` // New field
	MinPrice        int       `query:"minPrice"`
	MaxPrice        int       `query:"maxPrice"`
	Consensus       string    `query:"consensus"`      // New field
	MinScore        int       `query:"minScore"`       // New field
	MaxScore        int       `query:"maxScore"`       // New field
	MinCov          int       `query:"minCov"`         // New field
	MaxCov          int       `query:"maxCov"`         // New field
	MinCap          int64     `query:"minCap"`         // New field
	MaxCap          int64     `query:"maxCap"`         // New field
	MinVol          int64     `query:"minVol"`         // New field
	MaxVol          int64     `query:"maxVol"`         // New field
	MinOutstanding  int64     `query:"minOutstanding"` // New field
	MaxOutstanding  int64     `query:"maxOutstanding"` // New field
	MinBeta         int       `query:"minBeta"`        // New field
	MaxBeta         int       `query:"maxBeta"`        // New field
	MinEps          int       `query:"minEps"`         // New field
	MaxEps          int       `query:"maxEps"`         // New field
	MinPe           int       `query:"minPe"`          // New field
	MaxPe           int       `query:"maxPe"`          // New field
	Dividend        bool      `query:"dividend"`
	MinYield        int       `query:"minYield"`        // New field
	MaxYield        int       `query:"maxYield"`        // New field
	MinPayoutRatio  int       `query:"minPayoutRatio"`  // New field
	MaxPayoutRatio  int       `query:"maxPayoutRatio"`  // New field
	Frequency       []string  `query:"frequency"`       // New field
	MinHoldings     int       `query:"minHoldings"`     // New field
	MaxHoldings     int       `query:"maxHoldings"`     // New field
	Family          []string  `query:"family"`          // New field
	MinAum          int64     `query:"minAum"`          // New field
	MaxAum          int64     `query:"maxAum"`          // New field
	MinExpenseRatio int       `query:"minExpenseRatio"` // New field
	MaxExpenseRatio int       `query:"maxExpenseRatio"` // New field
	MinNav          int       `query:"minNav"`          // New field
	MaxNav          int       `query:"maxNav"`          // New field
	MinInception    time.Time `query:"minInception"`    // New field
	MaxInception    time.Time `query:"maxInception"`    // New field
	Order           []string  `query:"order"`
	Asc             string    `query:"asc"`
	Limit           int       `query:"limit"`
}

type SecParamsPointers struct {
	Exchange        *string  `query:"exchange"`
	Country         *string  `query:"country"`
	Currency        *string  `query:"currency"` // New field
	MinPrice        *float64 `query:"minPrice"`
	MaxPrice        *float64 `query:"maxPrice"`
	Consensus       *string  `query:"consensus"`      // New field
	MinScore        *float64 `query:"minScore"`       // New field
	MaxScore        *float64 `query:"maxScore"`       // New field
	MinCov          *int     `query:"minCov"`         // New field
	MaxCov          *int     `query:"maxCov"`         // New field
	MinCap          *string  `query:"minCap"`         // New field
	MaxCap          *string  `query:"maxCap"`         // New field
	MinVol          *string  `query:"minVol"`         // New field
	MaxVol          *string  `query:"maxVol"`         // New field
	MinOutstanding  *string  `query:"minOutstanding"` // New field
	MaxOutstanding  *string  `query:"maxOutstanding"` // New field
	MinBeta         *float64 `query:"minBeta"`        // New field
	MaxBeta         *float64 `query:"maxBeta"`        // New field
	MinEps          *float64 `query:"minEps"`         // New field
	MaxEps          *float64 `query:"maxEps"`         // New field
	MinPe           *float64 `query:"minPe"`          // New field
	MaxPe           *float64 `query:"maxPe"`          // New field
	Dividend        *bool    `query:"dividend"`
	MinYield        *float64 `query:"minYield"`        // New field
	MaxYield        *float64 `query:"maxYield"`        // New field - depends on dividend
	MinPayoutRatio  *float64 `query:"minPayoutRatio"`  // New field - depends on dividend
	MaxPayoutRatio  *float64 `query:"maxPayoutRatio"`  // New field - depends on dividend
	Frequency       *string  `query:"frequency"`       // New field - depends on dividend
	MinHoldings     *float64 `query:"minHoldings"`     // New field
	MaxHoldings     *float64 `query:"maxHoldings"`     // New field
	Family          *string  `query:"family"`          // New field
	MinAum          *string  `query:"minAum"`          // New field
	MaxAum          *string  `query:"maxAum"`          // New field
	MinExpenseRatio *float64 `query:"minExpenseRatio"` // New field
	MaxExpenseRatio *float64 `query:"maxExpenseRatio"` // New field
	MinNav          *float64 `query:"minNav"`          // New field
	MaxNav          *float64 `query:"maxNav"`          // New field
	MinInception    *string  `query:"minInception"`    // New field
	MaxInception    *string  `query:"maxInception"`    // New field
	Order           *string  `query:"order"`
	Asc             *string  `query:"asc"`
	Limit           *int     `query:"limit"`
}

// Possible valid values for Order and Asc fields
var (
	ValidOrderColumns = map[string]bool{
		"price":       true,
		"consensus":   true, // New field
		"score":       true, // New field
		"coverage":    true, // New field
		"marketcap":   true,
		"volume":      true,
		"avgvolume":   true,
		"outstanding": true, // New field
		"beta":        true, // New field
		"eps":         true, // New field
		"pe":          true, // New field
		"yield":       true, // New field
		"payout":      true, // New field
		"holdings":    true, // New field
		"aum":         true, // New field
		"expense":     true, // New field
		"nav":         true, // New field
		"inception":   true, // New field
		"pc":          true,
		"ppc":         true,
		"updated":     true,
	}
	ValidAscValues = map[string]bool{
		"asc":  true,
		"desc": true,
	}
)

func ValidateIntRange(min, max *float64, minout, maxout *int) error {
	// Validate numeric fields that become `int`
	var err error
	parseToInt := func(value *float64) (int, error) {
		if value == nil {
			return 0, nil
		}
		strValue := fmt.Sprintf("%.2f", *value)
		parsed, err := strconv.Atoi(helpers.NormalizeFloatStrToIntStr(strValue))
		if err != nil {
			return 0, fmt.Errorf("failed to convert value to int: %w", err)
		}
		return parsed, nil
	}

	*minout, err = parseToInt(min)
	if err != nil {
		return err
	}
	*maxout, err = parseToInt(max)
	if err != nil {
		return err
	}

	if min != nil && max != nil && *minout > *maxout {
		return errors.New("min cannot be greater than max")
	}

	return nil

}

func ValidateBigIntRange(min, max *string, minout, maxout *int64) error {
	// Validate numeric fields that become `int`
	var err error
	// Validate fields that become `int64`
	parseToInt64 := func(value *string) (int64, error) {
		if value == nil {
			return 0, nil
		}
		return helpers.ParseNumberString(*value)
	}

	*minout, err = parseToInt64(min)
	if err != nil {
		return err
	}
	*maxout, err = parseToInt64(max)
	if err != nil {
		return err
	}

	if min != nil && max != nil && *minout > *maxout {
		return errors.New("min cannot be greater than max")
	}

	return nil
}

// Validate method ensures that all fields are valid and normalizes them
func (p *SecParamsPointers) Validate() (*SecParams, error) {
	params := &SecParams{}

	// Helper function for parsing comma-separated strings
	parseCSV := func(value *string, upper bool) []string {
		if value == nil {
			return nil
		}
		var trimmed string
		if upper {
			trimmed = strings.ToUpper(strings.TrimSpace(*value))
		} else {
			trimmed = strings.ToLower(strings.TrimSpace(*value))
		}
		if trimmed == "" {
			return nil
		}
		return strings.Split(trimmed, ",")
	}

	// Validate and normalize string slice fields
	params.Exchange = parseCSV(p.Exchange, true)
	params.Country = parseCSV(p.Country, true)
	params.Currency = parseCSV(p.Currency, true)
	params.Order = parseCSV(p.Order, false)
	params.Family = parseCSV(p.Family, false)
	params.Frequency = parseCSV(p.Frequency, false)

	// Validate Consensus
	if p.Consensus != nil {
		*p.Consensus = strings.ToUpper(strings.TrimSpace(*p.Consensus))
		if *p.Consensus == "" {
			return nil, errors.New("consensus cannot be an empty string")
		}
		params.Consensus = *p.Consensus
	}

	// Validate MinCov and MaxCov
	if p.MinCov != nil && p.MaxCov != nil {
		if *p.MinCov > *p.MaxCov {
			return nil, errors.New("minCov cannot be greater than maxCov")
		}
		params.MinCov = *p.MinCov
		params.MaxCov = *p.MaxCov
	}

	// Validate MinPrice and MaxPrice
	err := ValidateIntRange(p.MinPrice, p.MaxPrice, &params.MinPrice, &params.MaxPrice)
	if err != nil {
		return nil, err
	}

	// Validate MinScore and MaxScore
	err = ValidateIntRange(p.MinScore, p.MaxScore, &params.MinScore, &params.MaxScore)
	if err != nil {
		return nil, err
	}
	// Validate MinBeta and MaxBeta
	err = ValidateIntRange(p.MinBeta, p.MaxBeta, &params.MinBeta, &params.MaxBeta)
	if err != nil {
		return nil, err
	}

	// Validate MinEps and MaxEps
	err = ValidateIntRange(p.MinEps, p.MaxEps, &params.MinEps, &params.MaxEps)
	if err != nil {
		return nil, err
	}
	// Validate MinPe and MaxPe
	err = ValidateIntRange(p.MinPe, p.MaxPe, &params.MinPe, &params.MaxPe)
	if err != nil {
		return nil, err
	}

	//Valudate MinHoldings and MaxHoldings
	err = ValidateIntRange(p.MinHoldings, p.MaxHoldings, &params.MinHoldings, &params.MaxHoldings)
	if err != nil {
		return nil, err
	}

	// Validate MinExpenseRatio and MaxExpenseRatio
	err = ValidateIntRange(p.MinExpenseRatio, p.MaxExpenseRatio, &params.MinExpenseRatio, &params.MaxExpenseRatio)
	if err != nil {
		return nil, err
	}

	// Validate MinNav and MaxNav
	err = ValidateIntRange(p.MinNav, p.MaxNav, &params.MinNav, &params.MaxNav)
	if err != nil {
		return nil, err
	}

	// Parse int64 fields
	err = ValidateBigIntRange(p.MinVol, p.MaxVol, &params.MinVol, &params.MaxVol)
	if err != nil {
		return nil, err
	}

	// Validate MinCap and MaxCap
	err = ValidateBigIntRange(p.MinCap, p.MaxCap, &params.MinCap, &params.MaxCap)
	if err != nil {
		return nil, err
	}

	// Validate MinOutstanding and MaxOutstanding
	err = ValidateBigIntRange(p.MinOutstanding, p.MaxOutstanding, &params.MinOutstanding, &params.MaxOutstanding)
	if err != nil {
		return nil, err
	}

	//Validate MinAum and MaxAum
	err = ValidateBigIntRange(p.MinAum, p.MaxAum, &params.MinAum, &params.MaxAum)
	if err != nil {
		return nil, err
	}

	// Validate MinInception and MaxInception (expect format YYYY-MM-DD)
	parseToDate := func(value *string) (time.Time, error) {
		if value == nil {
			return time.Time{}, nil
		}
		parsed, err := time.Parse("2006-01-02", *value)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date format for %s, expected YYYY-MM-DD", *value)
		}
		return parsed, nil
	}

	params.MinInception, err = parseToDate(p.MinInception)
	if err != nil {
		return nil, err
	}

	params.MaxInception, err = parseToDate(p.MaxInception)
	if err != nil {
		return nil, err
	}

	if !params.MinInception.IsZero() && !params.MaxInception.IsZero() && params.MinInception.After(params.MaxInception) {
		return nil, errors.New("minInception cannot be after maxInception")
	}

	// Validate Dividend and dependent fields
	if p.Dividend != nil {
		params.Dividend = *p.Dividend
	}

	// Validate MinYield and MaxYield

	if p.MinYield != nil {
		if !params.Dividend {
			return nil, errors.New("minYield and maxYield require dividend to be true")
		}

		err = ValidateIntRange(p.MinYield, p.MaxYield, &params.MinYield, &params.MaxYield)
		if err != nil {
			return nil, err
		}

	}

	// Validate MinPayoutRatio and MaxPayoutRatio

	if p.MinPayoutRatio != nil {
		if !params.Dividend {
			return nil, errors.New("minPayoutRatio and maxPayoutRatio require dividend to be true")
		}

		err = ValidateIntRange(p.MinPayoutRatio, p.MaxPayoutRatio, &params.MinPayoutRatio, &params.MaxPayoutRatio)
		if err != nil {
			return nil, err
		}
	}

	// Validate Order (must be a list of valid sorting fields)
	if p.Order != nil {
		orderList := parseCSV(p.Order, false)
		for _, col := range orderList {
			if !ValidOrderColumns[col] {
				return nil, fmt.Errorf("invalid order value: %s, must be one of %v", col, keys(ValidOrderColumns))
			}
		}
		params.Order = orderList
	}

	// Validate Asc
	if p.Asc != nil {
		*p.Asc = strings.ToLower(strings.TrimSpace(*p.Asc))
		if !ValidAscValues[*p.Asc] {
			return nil, fmt.Errorf("invalid asc value: %s, must be 'asc' or 'desc'", *p.Asc)
		}
		params.Asc = *p.Asc
	}

	// Validate Limit
	if p.Limit != nil {
		if *p.Limit <= 0 {
			return nil, errors.New("limit must be greater than 0")
		}
		params.Limit = *p.Limit
	}

	// All validations passed
	return params, nil
}

// Helper function to get the keys of a map
func keys(m map[string]bool) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}
