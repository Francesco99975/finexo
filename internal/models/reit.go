package models

import (
	"fmt"
	"strings"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// REIT represents a row from the reits table.
type REIT struct {
	Security   `json:"security"` // Embedded security properties
	Occupation NullableInt       `db:"occupation" json:"occupation,omitempty"`
	Focus      NullableString    `db:"focus" json:"focus,omitempty"`
	FFO        NullableInt       `db:"ffo" json:"ffo,omitempty"`
	PFFO       NullableInt       `db:"pffo" json:"pffo,omitempty"`
	Timing     NullableString    `db:"tm" json:"timing,omitempty"` // Enum: fwd, ttm
}

func (reit *REIT) flatten() map[string]any {
	return map[string]any{
		"ticker":     reit.Security.Ticker,
		"exchange":   reit.Security.Exchange,
		"occupation": reit.Occupation,
		"focus":      reit.Focus,
		"ffo":        reit.FFO,
		"pffo":       reit.PFFO,
		"tm":         reit.Timing,
	}
}

func (r *REIT) Scan(rows *sqlx.Rows) error {
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
		&r.Security.Ticker, &r.Security.Exchange, &r.Typology, &r.Currency, &r.FullName, &r.Sector, &r.Industry, &r.SubIndustry,
		&r.Price, &r.PC, &r.PCP, &r.YearLow, &r.YearHigh, &r.DayLow, &r.DayHigh, &r.Consensus, &r.Score, &r.Coverage,
		&r.MarketCap, &r.Volume, &r.AvgVolume, &r.Outstanding, &r.Beta,
		&r.PClose, &r.COpen, &r.Bid, &r.BidSize, &r.Ask, &r.AskSize,
		&r.EPS, &r.PE, &r.STM, &r.Created, &r.Updated,

		&r.Occupation, &r.Focus, &r.FFO, &r.PFFO, &r.Timing,

		// Dividend Fields
		&dividendYield, &dividendTiming, &dividendAnnualPayout, &dividendPayoutRatio,
		&dividendGrowthRate, &dividendYearsGrowth, &dividendLastAnnounced, &dividendFrequency,
		&dividendExDivDate, &dividendPayoutDate,
	)
	if err != nil {
		return err
	}

	// If dividend data exists, create the Dividend struct
	if dividendYield.Valid || dividendAnnualPayout.Valid || dividendPayoutRatio.Valid {
		r.Dividend = &Dividend{
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

func CreateReit(db *sqlx.DB, reit *REIT) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer database.HandleTransaction(tx, &err)

	// Insert into securities table
	if err := InsertSecurity(tx, &reit.Security); err != nil {
		return err
	}

	// Insert into reits table
	reitsQuery := `
		INSERT INTO reits (ticker, exchange, occupation, focus, ffo, pffo, tm)
		VALUES (:ticker, :exchange, :occupation, :focus, :ffo, :pffo, :tm)
	`
	_, err = tx.NamedExec(reitsQuery, reit.flatten())
	if err != nil {
		return fmt.Errorf("failed to insert reit: %w", err)
	}

	// Insert into dividends table if provided
	if err := InsertDividend(tx, reit.Security.Dividend); err != nil {
		return err
	}

	return nil
}

func UpdateREIT(db *sqlx.DB, reit *REIT) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer database.HandleTransaction(tx, &err)

	// Step 1: Update Security
	if err := UpdateSecurity(tx, &reit.Security); err != nil {
		return err
	}

	// Step 2: Update the REITs table
	query := "UPDATE reits SET "
	args := make(map[string]any)
	updates := []string{}

	// Nullable Integer Fields
	if reit.Occupation.Valid {
		updates = append(updates, "occupation = :occupation")
		args["occupation"] = reit.Occupation.Int64
	}
	if reit.FFO.Valid {
		updates = append(updates, "ffo = :ffo")
		args["ffo"] = reit.FFO.Int64
	}
	if reit.PFFO.Valid {
		updates = append(updates, "pffo = :pffo")
		args["pffo"] = reit.PFFO.Int64
	}

	// Nullable String Fields
	if reit.Focus.Valid {
		updates = append(updates, "focus = :focus")
		args["focus"] = reit.Focus.String
	}
	if reit.Timing.Valid {
		updates = append(updates, "tm = :timing")
		args["timing"] = reit.Timing.String
	}

	// Ensure at least one field is being updated
	if len(updates) > 0 {
		query += strings.Join(updates, ", ") + " WHERE ticker = :ticker AND exchange = :exchange"
		args["ticker"] = reit.Ticker
		args["exchange"] = reit.Exchange

		_, err = tx.NamedExec(query, args)
		if err != nil {
			return fmt.Errorf("failed to update REIT (%s): %w", reit.Ticker, err)
		}
	}

	// Step 3: Update Dividend
	if reit.Security.Dividend != nil {
		var divExists bool
		err = tx.Get(&divExists, "SELECT EXISTS(SELECT 1 FROM dividends WHERE ticker = $1 AND exchange = $2)", reit.Ticker, reit.Exchange)
		if err != nil {
			return fmt.Errorf("failed to check dividend existence: %w", err)
		}

		if divExists {
			if err := UpdateDividend(tx, reit.Security.Dividend); err != nil {
				return err
			}
		} else {
			if err := InsertDividend(tx, reit.Security.Dividend); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetREIT(db *sqlx.DB, input string) (*REIT, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected: ticker:exchange")
	}
	ticker, exchange := parts[0], parts[1]

	// Query for retrieving stock details, including dividend (if available)
	query := `
		SELECT
			s.ticker, s.exchange, s.typology, s.currency, s.fullname, s.sector, s.industry, s.subindustry,
			s.price, s.pc, s.pcp, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
			s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
			s.ask, s.asksz, s.eps, s.pe, s.stm, s.created, s.updated,

			r.occupation, r.focus, r.ffo, r.pffo, r.tm,

			d.yield AS dividend_yield, d.tm AS dividend_timing,
    d.ap AS dividend_annualPayout, d.pr AS dividend_payoutRatio,
    d.lgr AS dividend_growthRate, d.yog AS dividend_yearsGrowth,
    d.lad AS dividend_lastAnnounced, d.frequency AS dividend_frequency,
    d.edd AS dividend_exDivDate, d.pd AS dividend_payoutDate

		FROM securities s
		INNER JOIN reits r ON s.ticker = r.ticker AND s.exchange = r.exchange
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange AND s.typology = 'REIT'
	`

	// Execute the query using NamedQuery
	rows, err := db.NamedQuery(query, map[string]any{
		"ticker":   ticker,
		"exchange": exchange,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve reit '%s': %w", input, err)
	}
	defer rows.Close()

	// Check if any rows were returned
	if !rows.Next() {
		return nil, fmt.Errorf("reit '%s' not found", input)
	}

	// Parse result into Security struct
	var reit REIT
	if err := reit.Scan(rows); err != nil {
		return nil, fmt.Errorf("failed to scan reit '%s': %w", input, err)
	}

	return &reit, nil
}

func GetREITs(
	db *sqlx.DB,
	exchange, country []string,
	minPrice, maxPrice int,
	orderBy []string, orderDirection string,
	limit int,
	dividend bool,
) ([]REIT, error) {
	// Base query selecting relevant security fields where typology = 'STOCK'
	query := `
		SELECT
			s.ticker, s.exchange, s.typology, s.currency, s.fullname, s.sector, s.industry, s.subindustry,
			s.price, s.pc, s.pcp, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
			s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
			s.ask, s.asksz, s.eps, s.pe, s.stm, s.created, s.updated,

			r.occupation, r.focus, r.ffo, r.pffo, r.tm,

			d.yield AS dividend_yield, d.tm AS dividend_timing,
    d.ap AS dividend_annualPayout, d.pr AS dividend_payoutRatio,
    d.lgr AS dividend_growthRate, d.yog AS dividend_yearsGrowth,
    d.lad AS dividend_lastAnnounced, d.frequency AS dividend_frequency,
    d.edd AS dividend_exDivDate, d.pd AS dividend_payoutDate

		FROM securities s
		INNER JOIN reits r ON s.ticker = r.ticker AND s.exchange = r.exchange
	`

	// Adjust JOIN type based on dividend presence
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions
	query += `
	WHERE s.typology = 'REIT'
	AND (array_length(COALESCE($1, ARRAY[]::text[]), 1) = 0 OR s.exchange = ANY($1::text[]))
	AND (array_length(COALESCE($2, ARRAY[]::text[]), 1) = 0 OR s.exchange IN (SELECT title FROM exchanges WHERE cc = ANY($2::text[])))
	AND CAST($3 AS NUMERIC) = -1 OR s.price >= CAST($3 AS NUMERIC)
	AND CAST($4 AS NUMERIC) = -1 OR s.price <= CAST($4 AS NUMERIC)
`

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
		"pc":          "s.pc",
		"pcp":         "s.pcp",
		"updated":     "s.updated",
	}

	order := "s.price ASC" // Default ordering
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
		query += fmt.Sprintf(" LIMIT %d", limit)
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
		return nil, fmt.Errorf("failed to retrieve reits: %w", err)
	}
	defer rows.Close()

	// Parse results
	var reits []REIT
	for rows.Next() {
		var reit REIT
		if err := reit.Scan(rows); err != nil {
			return nil, fmt.Errorf("failed to scan reit row: %w", err)
		}
		reits = append(reits, reit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over reit rows: %w", err)
	}

	return reits, nil
}
