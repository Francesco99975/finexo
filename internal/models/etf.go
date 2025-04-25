package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ETF represents a row from the etfs table.
type ETF struct {
	Security          `json:"security"` // Embedded security properties
	Holdings          int               `db:"holdings" json:"holdings"`
	Family            string            `db:"family" json:"family"`
	AUM               NullableInt       `db:"aum" json:"aum,omitempty"`
	ExpenseRatio      NullableInt       `db:"er" json:"expenseRatio,omitempty"`
	NAV               NullableInt       `db:"nav" json:"nav,omitempty"`
	InceptionDate     NullableTime      `db:"inception" json:"inception,omitempty"`
	RelatedSecurities []string          `json:"relatedSecurities"` // Related securities as "TICKER:EXCHANGE:ALLOCATION"
}

func (etf *ETF) PrettyPrintString() string {
	var sb strings.Builder

	sb.WriteString(etf.Security.CreatePrettyPrintString())

	sb.WriteString("Holdings: " + strconv.Itoa(etf.Holdings) + " -- ")
	sb.WriteString("Family: " + etf.Family + " -- ")
	if etf.AUM.Valid {
		sb.WriteString("AUM: " + strconv.Itoa(int(etf.AUM.Int64)) + " -- ")
	}
	if etf.ExpenseRatio.Valid {
		sb.WriteString("Expense Ratio: " + strconv.Itoa(int(etf.ExpenseRatio.Int64)) + " -- ")
	}
	if etf.NAV.Valid {
		sb.WriteString("NAV: " + strconv.Itoa(int(etf.NAV.Int64)) + " -- ")
	}
	if etf.InceptionDate.Valid {
		sb.WriteString("Inception Date: " + etf.InceptionDate.Time.Format("2006-01-02") + " -- ")
	}
	if len(etf.RelatedSecurities) > 0 {
		sb.WriteString("Related Securities: " + strings.Join(etf.RelatedSecurities, ", ") + " -- ")
	}
	return sb.String()
}

func (etf *ETF) flatten() map[string]any {
	return map[string]any{
		"ticker":       etf.Security.Ticker,
		"exchange":     etf.Security.Exchange,
		"family":       etf.Family,
		"holdings":     etf.Holdings,
		"aum":          etf.AUM,
		"expenseRatio": etf.ExpenseRatio,
		"nav":          etf.NAV,
		"inception":    etf.InceptionDate,
	}
}

func (etf *ETF) Scan(rows *sqlx.Rows) error {
	// Temporary variables for nullable fields
	var relatedSecurities NullableString

	// Define variables to scan values
	var (
		dividendYield         NullableInt
		dividendTiming        NullableString
		dividendAnnualPayout  NullableInt
		dividendPayoutRatio   NullableInt
		dividendGrowthRate    NullableInt
		dividendYearsGrowth   NullableInt
		dividendLastAnnounced NullableInt
		dividendFrequency     NullableString
		dividendExDivDate     NullableTime
		dividendPayoutDate    NullableTime
	)

	// Scan all fields from the row
	err := rows.Scan(
		&etf.Security.Ticker,
		&etf.Security.Exchange,
		&etf.Security.Typology,
		&etf.Security.Currency,
		&etf.Security.FullName,
		&etf.Security.Sector,
		&etf.Security.Industry,
		&etf.Security.SubIndustry,
		&etf.Security.Price,
		&etf.Security.PC,
		&etf.Security.PCP,
		&etf.Security.YearLow,
		&etf.Security.YearHigh,
		&etf.Security.DayLow,
		&etf.Security.DayHigh,
		&etf.Security.Consensus,
		&etf.Security.Score,
		&etf.Security.Coverage,
		&etf.Security.MarketCap,
		&etf.Security.Volume,
		&etf.Security.AvgVolume,
		&etf.Security.Outstanding,
		&etf.Security.Beta,
		&etf.Security.PClose,
		&etf.Security.COpen,
		&etf.Security.Bid,
		&etf.Security.BidSize,
		&etf.Security.Ask,
		&etf.Security.AskSize,
		&etf.Security.EPS,
		&etf.Security.PE,
		&etf.Security.Target,
		&etf.Security.STM,
		&etf.Security.Created,
		&etf.Security.Updated,

		// ETF-specific fields (order fixed)
		&etf.Holdings,
		&etf.Family,
		&etf.AUM,
		&etf.ExpenseRatio,
		&etf.NAV,
		&etf.InceptionDate,

		// Related Securities
		&relatedSecurities,

		// Dividend Fields
		&dividendYield, &dividendTiming, &dividendAnnualPayout, &dividendPayoutRatio,
		&dividendGrowthRate, &dividendYearsGrowth, &dividendLastAnnounced, &dividendFrequency,
		&dividendExDivDate, &dividendPayoutDate,
	)

	if err != nil {
		return fmt.Errorf("failed to scan ETF fields: %w", err)
	}

	if relatedSecurities.Valid {
		etf.RelatedSecurities = strings.Split(strings.ReplaceAll(relatedSecurities.String, "|", ":"), ",")
	} else {
		etf.RelatedSecurities = []string{}
	}

	// If dividend data exists, create the Dividend struct
	if dividendYield.Valid || dividendAnnualPayout.Valid || dividendPayoutRatio.Valid {
		etf.Security.Dividend = &Dividend{
			Yield:         int(dividendYield.Int64),
			Timing:        dividendTiming,
			AnnualPayout:  dividendAnnualPayout,
			PayoutRatio:   dividendPayoutRatio,
			GrowthRate:    dividendGrowthRate,
			YearsGrowth:   dividendYearsGrowth,
			LastAnnounced: dividendLastAnnounced,
			Frequency:     dividendFrequency,
			ExDivDate:     dividendExDivDate,
			PayoutDate:    dividendPayoutDate,
		}
	}

	return nil
}

func CreateETF(db *sqlx.DB, etf *ETF) error {
	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer database.HandleTransaction(tx, &err)

	// Step 1: Insert Security (Reused function)
	if err := InsertSecurity(tx, &etf.Security); err != nil {
		return err
	}

	// Step 2: Insert into the etfs table
	etfsQuery := `
		INSERT INTO etfs (ticker, exchange, family, holdings, aum, er, nav, inception)
		VALUES (:ticker, :exchange, :family, :holdings, :aum, :expenseRatio, :nav, :inception)
	`
	_, err = tx.NamedExec(etfsQuery, etf.flatten())
	if err != nil {
		return fmt.Errorf("failed to insert ETF: %w", err)
	}

	// Step 3: Insert related securities into etf_related_securities
	if len(etf.RelatedSecurities) > 0 {
		relatedQuery := `
			INSERT INTO etf_related_securities (etf_ticker, etf_exchange, related_ticker, related_exchange, allocation)
			VALUES (:etf_ticker, :etf_exchange, :related_ticker, :related_exchange, :allocation)
		`

		for _, related := range etf.RelatedSecurities {
			parts := strings.Split(related, ":")
			if len(parts) != 3 {
				return fmt.Errorf("invalid related security format (expected TICKER:EXCHANGE:ALLOCATION): %s", related)
			}

			allocation, err := strconv.Atoi(parts[2]) // Convert allocation to int
			if err != nil {
				return fmt.Errorf("invalid allocation value: %s", parts[2])
			}

			_, err = tx.NamedExec(relatedQuery, map[string]any{
				"etf_ticker":       etf.Ticker,
				"etf_exchange":     etf.Exchange,
				"related_ticker":   parts[0],   // Ticker
				"related_exchange": parts[1],   // Exchange
				"allocation":       allocation, // Allocation as int
			})
			if err != nil {
				return fmt.Errorf("failed to insert related security '%s': %w", related, err)
			}
		}
	}

	// Step 4: Insert Dividend (Reused function)
	if err := InsertDividend(tx, etf.Security.Dividend); err != nil {
		return err
	}

	return nil
}

func UpdateETF(db *sqlx.DB, etf *ETF) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer database.HandleTransaction(tx, &err)

	// Step 1: Update Security
	if err := UpdateSecurity(tx, &etf.Security); err != nil {
		return err
	}

	// Step 2: Update the ETFs table
	query := "UPDATE etfs SET "
	args := make(map[string]any)
	updates := []string{}

	if etf.Family != "" {
		updates = append(updates, "family = :family")
		args["family"] = etf.Family
	}
	if etf.Holdings != 0 {
		updates = append(updates, "holdings = :holdings")
		args["holdings"] = etf.Holdings
	}
	if etf.AUM.Valid {
		updates = append(updates, "aum = :aum")
		args["aum"] = etf.AUM.Int64
	}
	if etf.ExpenseRatio.Valid {
		updates = append(updates, "er = :expenseRatio")
		args["expenseRatio"] = etf.ExpenseRatio.Int64
	}
	if etf.NAV.Valid {
		updates = append(updates, "nav = :nav")
		args["nav"] = etf.NAV.Int64
	}
	if etf.InceptionDate.Valid {
		updates = append(updates, "inception = :inception")
		args["inception"] = etf.InceptionDate.Time
	}

	// Ensure at least one field is being updated
	if len(updates) > 0 {
		query += strings.Join(updates, ", ") + " WHERE ticker = :ticker AND exchange = :exchange"
		args["ticker"] = etf.Ticker
		args["exchange"] = etf.Exchange

		_, err = tx.NamedExec(query, args)
		if err != nil {
			return fmt.Errorf("failed to update ETF (%s): %w", etf.Ticker, err)
		}
	}

	// Step 3: Update ETF Related Securities
	if len(etf.RelatedSecurities) > 0 {
		// First, delete existing related securities
		_, err = tx.Exec("DELETE FROM etf_related_securities WHERE etf_ticker = $1 AND etf_exchange = $2", etf.Ticker, etf.Exchange)
		if err != nil {
			return fmt.Errorf("failed to delete existing related securities for %s: %w", etf.Ticker, err)
		}

		// Reinsert new related securities
		relatedQuery := `
			INSERT INTO etf_related_securities (etf_ticker, etf_exchange, related_ticker, related_exchange, allocation)
			VALUES (:etf_ticker, :etf_exchange, :related_ticker, :related_exchange, :allocation)
		`

		for _, related := range etf.RelatedSecurities {
			parts := strings.Split(related, ":")
			if len(parts) != 3 {
				return fmt.Errorf("invalid related security format (expected TICKER:EXCHANGE:ALLOCATION): %s", related)
			}

			allocation, err := strconv.Atoi(parts[2]) // Convert allocation to int
			if err != nil {
				return fmt.Errorf("invalid allocation value: %s", parts[2])
			}

			_, err = tx.NamedExec(relatedQuery, map[string]any{
				"etf_ticker":       etf.Ticker,
				"etf_exchange":     etf.Exchange,
				"related_ticker":   parts[0],   // Ticker
				"related_exchange": parts[1],   // Exchange
				"allocation":       allocation, // Allocation as int
			})
			if err != nil {
				return fmt.Errorf("failed to insert related security '%s': %w", related, err)
			}
		}
	}

	// Step 4: Update Dividend
	if etf.Security.Dividend != nil {
		var divExists bool
		err = tx.Get(&divExists, "SELECT EXISTS(SELECT 1 FROM dividends WHERE ticker = $1 AND exchange = $2)", etf.Ticker, etf.Exchange)
		if err != nil {
			return fmt.Errorf("failed to check dividend existence: %w", err)
		}

		if divExists {
			if err := UpdateDividend(tx, etf.Security.Dividend); err != nil {
				return err
			}
		} else {
			if err := InsertDividend(tx, etf.Security.Dividend); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetETF(db *sqlx.DB, input string) (*ETF, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected '<ticker>:<exchange>'")
	}
	ticker, exchange := parts[0], parts[1]

	// Query for retrieving ETF details, including dividend (if available)
	query := `
		SELECT
			s.ticker, s.exchange, s.typology, s.currency, s.fullname, s.sector, s.industry, s.subindustry,
			s.price, s.pc, s.pcp, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
			s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
			s.ask, s.asksz, s.eps, s.pe, s.target, s.stm, s.created, s.updated,

		e.holdings, e.family, e.aum, e.er AS expenseRatio, e.nav, e.inception,


			COALESCE(STRING_AGG(
					er.related_ticker || '|' || er.related_exchange || '|' || er.allocation, ','
			), '') AS related_securities,

			d.yield AS dividend_yield, d.tm AS dividend_timing,
    d.ap AS dividend_annualPayout, d.pr AS dividend_payoutRatio,
    d.lgr AS dividend_growthRate, d.yog AS dividend_yearsGrowth,
    d.lad AS dividend_lastAnnounced, d.frequency AS dividend_frequency,
    d.edd AS dividend_exDivDate, d.pd AS dividend_payoutDate

		FROM securities s
		INNER JOIN etfs e ON s.ticker = e.ticker AND s.exchange = e.exchange
		LEFT JOIN etf_related_securities er ON e.ticker = er.etf_ticker AND e.exchange = er.etf_exchange
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange AND s.typology = 'ETF'
		GROUP BY s.ticker, s.exchange, e.holdings, e.family, e.aum, e.er, e.nav, e.inception, d.yield, d.tm, d.ap, d.pr, d.lgr, d.yog, d.lad, d.frequency, d.edd, d.pd
	`

	// Execute the query
	rows, err := db.NamedQuery(query, map[string]any{
		"ticker":   ticker,
		"exchange": exchange,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve etf '%s': %w", input, err)
	}
	defer rows.Close()

	// Check if any rows were returned
	if !rows.Next() {
		return nil, fmt.Errorf("etf '%s' not found", input)
	}

	// Parse result into Security struct
	var etf ETF
	if err := etf.Scan(rows); err != nil {
		return nil, fmt.Errorf("failed to scan etf '%s': %w", input, err)
	}

	return &etf, nil
}

func GetETFs(
	db *sqlx.DB,
	exchange, country []string,
	minPrice, maxPrice int,
	orderBy []string, orderDirection string,
	limit int,
	dividend bool,
) ([]ETF, error) {
	// Base query selecting relevant ETF fields, including related securities
	query := `
		SELECT
			s.ticker, s.exchange, s.typology, s.currency, s.fullname, s.sector, s.industry, s.subindustry,
			s.price, s.pc, s.pcp, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
			s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
			s.ask, s.asksz, s.eps, s.pe, s.target, s.stm, s.created, s.updated,

			e.holdings, e.family, e.aum, e.er AS expenseRatio, e.nav, e.inception,


			COALESCE(STRING_AGG(
					er.related_ticker || '|' || er.related_exchange || '|' || er.allocation, ','
			), '') AS related_securities,


			d.yield AS dividend_yield, d.tm AS dividend_timing,
			d.ap AS dividend_annualPayout, d.pr AS dividend_payoutRatio,
			d.lgr AS dividend_growthRate, d.yog AS dividend_yearsGrowth,
			d.lad AS dividend_lastAnnounced, d.frequency AS dividend_frequency,
			d.edd AS dividend_exDivDate, d.pd AS dividend_payoutDate

		FROM securities s
		INNER JOIN etfs e ON s.ticker = e.ticker AND s.exchange = e.exchange
		LEFT JOIN etf_related_securities er ON e.ticker = er.etf_ticker AND e.exchange = er.etf_exchange
	`

	// Adjust JOIN type based on dividend presence
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions
	query += `
		WHERE s.typology = 'ETF'
		AND (array_length(COALESCE($1, ARRAY[]::text[]), 1) = 0 OR s.exchange = ANY($1::text[]))
		AND (array_length(COALESCE($2, ARRAY[]::text[]), 1) = 0 OR s.exchange IN (SELECT title FROM exchanges WHERE cc = ANY($2::text[])))
		AND CAST($3 AS NUMERIC) = -1 OR s.price >= CAST($3 AS NUMERIC)
		AND CAST($4 AS NUMERIC) = -1 OR s.price <= CAST($4 AS NUMERIC)
	`

	// Grouping by ETF to ensure `STRING_AGG()` works correctly
	query += " GROUP BY s.ticker, s.exchange, e.holdings, e.family, e.aum, e.er, e.nav, e.inception, d.yield, d.tm, d.ap, d.pr, d.lgr, d.yog, d.lad, d.frequency, d.edd, d.pd"

	// Apply ordering
	orderColumn := map[string]string{
		"price":       "s.price",
		"consensus":   "s.consensus", // New field
		"score":       "s.score",     // New field
		"coverage":    "s.coverage",  // New field
		"volume":      "s.volume",
		"avgvolume":   "s.avgvolume",
		"marketcap":   "s.cap",
		"outstanding": "s.outstanding", // New field
		"beta":        "s.beta",        // New field
		"eps":         "s.eps",         // New field
		"pe":          "s.pe",          // New field
		"yield":       "d.yield",       // New field
		"payout":      "d.pr",          // New field
		"holdings":    "e.holdings",    // New field
		"aum":         "e.aum",         // New field
		"expense":     "e.er",          // New field
		"nav":         "e.nav",         // New field
		"inception":   "e.inception",   // New field
		"pc":          "s.pc",
		"pcp":         "s.pcp",
		"updated":     "s.updated",
	}

	order := "s.price ASC"
	if len(orderBy) > 0 {
		for _, col := range orderBy {
			if colx, exists := orderColumn[col]; exists {
				if orderDirection == "desc" {
					order = fmt.Sprintf("%s DESC", colx)
				} else {
					order = fmt.Sprintf("%s ASC", colx)
				}
			}
		}
	}
	query += fmt.Sprintf(" ORDER BY %s", order)

	// PostgreSQL does NOT support named parameters in LIMIT
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit) // Convert to integer before execution
	}

	// Handle empty slices by converting them to PostgreSQL-friendly empty arrays
	var exchangeArray any = "{}" // PostgreSQL empty array format
	var countryArray any = "{}"

	if len(exchange) > 0 {
		exchangeArray = pq.Array(exchange) // Use pq.Array() only when non-empty
	}
	if len(country) > 0 {
		countryArray = pq.Array(country)
	}

	// Define "unset" sentinel values
	const unsetPrice = -1

	// If minPrice or maxPrice are 0, replace with -1
	minP := minPrice
	maxP := maxPrice
	if minPrice == 0 {
		minP = unsetPrice
	}
	if maxPrice == 0 {
		maxP = unsetPrice
	}

	args := []any{
		exchangeArray, // Wrap slices in pq.Array() for PostgreSQL compatibility
		countryArray,
		minP, // Now it defaults to -1 instead of 0
		maxP,
	}

	// Execute query using `Queryx`, NOT `NamedQuery`
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ETFs rows: %w", err)
	}
	defer rows.Close()

	// Parse results
	var etfs []ETF
	for rows.Next() {
		var etf ETF
		if err := etf.Scan(rows); err != nil { // Use custom Scan() method
			return nil, fmt.Errorf("failed to scan ETF row: %w", err)
		}
		etfs = append(etfs, etf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over ETF rows: %w", err)
	}

	return etfs, nil
}
