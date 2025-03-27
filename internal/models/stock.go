package models

import (
	"fmt"
	"strings"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func CreateStock(db *sqlx.DB, stock *Security) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer database.HandleTransaction(tx, &err)

	if err := InsertSecurity(tx, stock); err != nil {
		return err
	}

	if err := InsertDividend(tx, stock.Dividend); err != nil {
		return err
	}

	return nil
}

func UpdateStock(db *sqlx.DB, stock *Security) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer database.HandleTransaction(tx, &err)

	// Check if stock exists
	var exists bool
	err = tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM securities WHERE ticker = $1 AND exchange = $2)", stock.Ticker, stock.Exchange)
	if err != nil {
		return fmt.Errorf("failed to check stock existence: %w", err)
	}

	// If stock exists, update it; otherwise, insert it
	if exists {
		if err := UpdateSecurity(tx, stock); err != nil {
			return err
		}
	} else {
		if err := InsertSecurity(tx, stock); err != nil {
			return err
		}
	}

	// Handle Dividend update/insert
	if stock.Dividend != nil {
		var divExists bool
		err = tx.Get(&divExists, "SELECT EXISTS(SELECT 1 FROM dividends WHERE ticker = $1 AND exchange = $2)", stock.Dividend.Ticker, stock.Dividend.Exchange)
		if err != nil {
			return fmt.Errorf("failed to check dividend existence: %w", err)
		}

		if divExists {
			if err := UpdateDividend(tx, stock.Dividend); err != nil {
				return err
			}
		} else {
			if err := InsertDividend(tx, stock.Dividend); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetStock(db *sqlx.DB, input string) (*Security, error) {
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

			d.yield AS dividend_yield, d.tm AS dividend_timing,
    d.ap AS dividend_annualPayout, d.pr AS dividend_payoutRatio,
    d.lgr AS dividend_growthRate, d.yog AS dividend_yearsGrowth,
    d.lad AS dividend_lastAnnounced, d.frequency AS dividend_frequency,
    d.edd AS dividend_exDivDate, d.pd AS dividend_payoutDate

		FROM securities s
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange AND s.typology = 'STOCK'
	`

	// Execute the query using NamedQuery
	rows, err := db.NamedQuery(query, map[string]any{
		"ticker":   ticker,
		"exchange": exchange,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stock '%s': %w", input, err)
	}
	defer rows.Close()

	// Check if any rows were returned
	if !rows.Next() {
		return nil, fmt.Errorf("stock '%s' not found", input)
	}

	// Parse result into Security struct
	var stock Security
	if err := stock.Scan(rows); err != nil {
		return nil, fmt.Errorf("failed to scan stock '%s': %w", input, err)
	}

	return &stock, nil
}

func GetStocks(
	db *sqlx.DB,
	exchange, country []string,
	minPrice, maxPrice int,
	orderBy []string, orderDirection string,
	limit int,
	dividend bool,
) ([]Security, error) {
	// Base query selecting relevant security fields where typology = 'STOCK'
	query := `
		SELECT
    	s.ticker, s.exchange, s.typology, s.currency, s.fullname, s.sector, s.industry, s.subindustry,
		s.price, s.pc, s.pcp, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
		s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
		s.ask, s.asksz, s.eps, s.pe, s.stm, s.created, s.updated,

    	d.yield AS dividend_yield, d.tm AS dividend_timing,
    	d.ap AS dividend_annualPayout, d.pr AS dividend_payoutRatio,
    	d.lgr AS dividend_growthRate, d.yog AS dividend_yearsGrowth,
    	d.lad AS dividend_lastAnnounced, d.frequency AS dividend_frequency,
    	d.edd AS dividend_exDivDate, d.pd AS dividend_payoutDate
		FROM securities s
	`

	// Adjust JOIN type based on dividend presence
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions (Switch from named parameters to positional `$1, $2, etc.`)
	query += `
	WHERE s.typology = 'STOCK'
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
		return nil, fmt.Errorf("failed to retrieve stocks: %w", err)
	}
	defer rows.Close()

	// Parse results
	var stocks []Security
	for rows.Next() {
		var stock Security
		if err := stock.Scan(rows); err != nil {
			return nil, fmt.Errorf("failed to scan stock row: %w", err)
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over stock rows: %w", err)
	}

	return stocks, nil
}
