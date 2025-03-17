package models

import (
	"fmt"
	"strings"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/jmoiron/sqlx"
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
	exchange, country string,
	minPrice, maxPrice int,
	orderBy, orderDirection string,
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

	// WHERE conditions
	query += `
	WHERE s.typology = 'STOCK'
	AND (COALESCE(:exchange, '') = '' OR s.exchange = CAST(:exchange AS TEXT))
	AND (COALESCE(:country, '') = '' OR s.exchange IN (SELECT title FROM exchanges WHERE cc = :country))
	AND (COALESCE(:minPrice, -1) = -1 OR s.price >= CAST(:minPrice AS NUMERIC))
	AND (COALESCE(:maxPrice, -1) = -1 OR s.price <= CAST(:maxPrice AS NUMERIC))
`

	// Apply ordering
	orderColumn := map[string]string{
		"price":     "s.price",
		"volume":    "s.volume",
		"avgvolume": "s.avgvolume",
		"marketcap": "s.cap",
		"pc":        "s.pc",
		"pcp":       "s.pcp",
		"updated":   "s.updated",
	}

	order := "s.price ASC" // Default ordering
	if orderBy != "" {
		if col, exists := orderColumn[orderBy]; exists {
			if orderDirection == "desc" {
				order = fmt.Sprintf("%s DESC", col)
			} else {
				order = fmt.Sprintf("%s ASC", col)
			}
		}
	}
	query += fmt.Sprintf(" ORDER BY %s", order)

	// PostgreSQL does NOT support named parameters in LIMIT
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	// Query parameters (use sql.NullString and sql.NullInt64 for explicit types)
	params := map[string]any{
		"exchange": NullableString{String: exchange, Valid: exchange != ""},
		"country":  NullableString{String: country, Valid: country != ""},
		"minPrice": NullableInt{Int64: int64(minPrice), Valid: minPrice > 0},
		"maxPrice": NullableInt{Int64: int64(maxPrice), Valid: maxPrice > 0},
	}

	// log.Debugf("Executing Query: %s", query)
	// log.Debugf("Params: %+v", params)

	// Execute the query
	rows, err := db.NamedQuery(query, params)
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
