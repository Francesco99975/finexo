package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
)

// Security represents a row from the securities table.
type Security struct {
	Ticker      string         `db:"ticker" json:"ticker"`
	Country     string         `db:"cc" json:"country"`
	Suffix      string         `db:"suffix" json:"suffix"`
	Exchange    string         `db:"exchange" json:"exchange"`
	Typology    string         `db:"typology" json:"typology"`
	Fullname    string         `db:"fullname" json:"fullname"`
	Price       int            `db:"price" json:"price"`
	PriceChange int            `db:"pc" json:"priceChange"`
	PricePct    string         `db:"ppc" json:"pricePct"`
	YearRange   string         `db:"yrange" json:"yearRange"`
	DayRange    string         `db:"drange" json:"dayRange"`
	MarketCap   sql.NullString `db:"marketcap" json:"marketCap,omitempty"`
	Volume      sql.NullInt64  `db:"volume" json:"volume,omitempty"`
	AvgVolume   sql.NullInt64  `db:"avgvlm" json:"avgVolume,omitempty"`
	Beta        sql.NullString `db:"beta" json:"beta,omitempty"`
	PClose      int            `db:"pclose" json:"pclose"`
	COpen       int            `db:"copen" json:"copen"`
	Bid         int            `db:"bid" json:"bid"`
	BidSize     sql.NullInt64  `db:"bidsz" json:"bidSize,omitempty"`
	Ask         int            `db:"ask" json:"ask"`
	AskSize     sql.NullInt64  `db:"asksz" json:"askSize,omitempty"`
	Currency    string         `db:"currency" json:"currency"`
	Created     time.Time      `db:"created" json:"created"`
	Updated     time.Time      `db:"updated" json:"updated"`

	Dividend *Dividend `db:"-" json:"dividend,omitempty"` // Associated dividend data (if exists)
}

// Scan implements the sql.Scanner interface for Security.
func (s *Security) Scan(src any) error {
	// Convert the source into the appropriate structure.
	switch data := src.(type) {
	case map[string]any:
		// Manually map fields to the Security struct
		s.Ticker = data["ticker"].(string)
		s.Country = data["cc"].(string)
		s.Suffix = data["suffix"].(string)
		s.Exchange = data["exchange"].(string)
		s.Typology = data["typology"].(string)
		s.Fullname = data["fullname"].(string)
		s.Price = int(data["price"].(int64))
		s.PriceChange = int(data["pc"].(int64))
		s.PricePct = data["ppc"].(string)
		s.YearRange = data["yrange"].(string)
		s.DayRange = data["drange"].(string)

		// Handle nullable fields
		if marketcap, ok := data["marketcap"].(string); ok {
			s.MarketCap = sql.NullString{String: marketcap, Valid: true}
		}
		if volume, ok := data["volume"].(int64); ok {
			s.Volume = sql.NullInt64{Int64: volume, Valid: true}
		}

		// Handle related Dividend data
		if dividendData, ok := data["dividend"].(map[string]any); ok {
			dividend := &Dividend{}
			dividend.Ticker = dividendData["ticker"].(string)
			dividend.Exchange = dividendData["exchange"].(string)
			dividend.Rate = int(dividendData["rate"].(int64))
			s.Dividend = dividend
		}

	default:
		return fmt.Errorf("unsupported Scan source: %T", src)
	}

	return nil
}

// Dividend represents a row from the dividends table.
type Dividend struct {
	Ticker        string         `db:"ticker" json:"ticker"`
	Exchange      string         `db:"cc" json:"country"`
	Rate          int            `db:"rate" json:"rate"`
	RateType      string         `db:"trate" json:"rateType"`
	Yield         string         `db:"yield" json:"yield"`
	YieldType     string         `db:"tyield" json:"yieldType"`
	AnnualPayout  sql.NullInt64  `db:"ap" json:"annualPayout,omitempty"`
	APType        sql.NullString `db:"tap" json:"apType,omitempty"`
	PayoutRatio   sql.NullString `db:"pr" json:"payoutRatio,omitempty"`
	GrowthRate    sql.NullString `db:"lgr" json:"growthRate,omitempty"`
	YearsGrowth   sql.NullInt64  `db:"yog" json:"yearsGrowth,omitempty"`
	LastAnnounced sql.NullInt64  `db:"lad" json:"lastAnnounced,omitempty"`
	Frequency     sql.NullString `db:"frequency" json:"frequency,omitempty"`
	ExDivDate     sql.NullTime   `db:"edd" json:"exDivDate,omitempty"`
	PayoutDate    sql.NullTime   `db:"pd" json:"payoutDate,omitempty"`
}

// Stock represents a row from the stocks table.
type Stock struct {
	Security `json:"security"` // Embedded security properties
	EPS      sql.NullString    `db:"eps" json:"eps,omitempty"`
	EPSType  sql.NullString    `db:"teps" json:"epsType,omitempty"`
	PE       sql.NullString    `db:"pe" json:"pe,omitempty"`
	PEType   sql.NullString    `db:"tpe" json:"peType,omitempty"`
}

// ETF represents a row from the etfs table.
type ETF struct {
	Security          `json:"security"` // Embedded security properties
	Holdings          int               `db:"holdings" json:"holdings"`
	AUM               sql.NullString    `db:"aum" json:"aum,omitempty"`
	ExpenseRatio      sql.NullString    `db:"er" json:"expenseRatio,omitempty"`
	RelatedSecurities []string          `json:"relatedSecurities"` // Related securities for the ETF
}

func (etf *ETF) Scan(rows *sqlx.Rows) error {
	// Temporary variables for related securities and nullable fields
	var relatedSecurities string

	// Scan all fields
	err := rows.Scan(
		&etf.Security.Ticker,
		&etf.Security.Country,
		&etf.Security.Suffix,
		&etf.Security.Exchange,
		&etf.Security.Typology,
		&etf.Security.Fullname,
		&etf.Security.Price,
		&etf.Security.PriceChange,
		&etf.Security.PricePct,
		&etf.Security.YearRange,
		&etf.Security.DayRange,
		&etf.Security.MarketCap,
		&etf.Security.Volume,
		&etf.Security.AvgVolume,
		&etf.Security.Beta,
		&etf.Security.PClose,
		&etf.Security.COpen,
		&etf.Security.Bid,
		&etf.Security.BidSize,
		&etf.Security.Ask,
		&etf.Security.AskSize,
		&etf.Security.Currency,
		&etf.Security.Created,
		&etf.Security.Updated,
		&etf.Holdings,
		&etf.AUM,
		&etf.ExpenseRatio,
		&relatedSecurities, // Comma-separated related securities
		&etf.Security.Dividend.Rate,
		&etf.Security.Dividend.RateType,
		&etf.Security.Dividend.Yield,
		&etf.Security.Dividend.YieldType,
		&etf.Security.Dividend.AnnualPayout,
		&etf.Security.Dividend.APType,
		&etf.Security.Dividend.PayoutRatio,
		&etf.Security.Dividend.GrowthRate,
		&etf.Security.Dividend.YearsGrowth,
		&etf.Security.Dividend.LastAnnounced,
		&etf.Security.Dividend.Frequency,
		&etf.Security.Dividend.ExDivDate,
		&etf.Security.Dividend.PayoutDate,
	)
	if err != nil {
		return fmt.Errorf("failed to scan ETF fields: %w", err)
	}

	// Parse related securities
	if relatedSecurities != "" {
		etf.RelatedSecurities = strings.Split(relatedSecurities, ",")
	} else {
		etf.RelatedSecurities = []string{}
	}

	return nil
}

// REIT represents a row from the reits table.
type REIT struct {
	Security `json:"security"` // Embedded security properties
	FFO      sql.NullInt64     `db:"ffo" json:"ffo,omitempty"`
	FFOType  sql.NullString    `db:"tffo" json:"ffoType,omitempty"`
	PFFO     sql.NullInt64     `db:"pffo" json:"pffo,omitempty"`
	PFFOType sql.NullString    `db:"tpffo" json:"pffoType,omitempty"`
}

func CreateStock(db *sqlx.DB, stock *Stock) error {
	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback()
			if err != nil {
				log.Errorf("Failed to rollback transaction: %v\n", err)
			}
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Errorf("Failed to rollback transaction: %v\n", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Errorf("Failed to commit transaction: %v\n", err)
			}
		}
	}()

	// Insert into securities table
	securitiesQuery := `
		INSERT INTO securities (
			ticker, cc, suffix, exchange, typology, fullname, price, pc, ppc,
			yrange, drange, marketcap, volume, avgvlm, beta, pclose, copen, bid, bidsz,
			ask, asksz, currency, created, updated
		) VALUES (
			:ticker, :cc, :suffix, :exchange, 'stock', :fullname, :price, :pc, :ppc,
			:yrange, :drange, :marketcap, :volume, :avgvlm, :beta, :pclose, :copen, :bid, :bidsz,
			:ask, :asksz, :currency, :created, :updated
		)
	`
	_, err = tx.NamedExec(securitiesQuery, &stock.Security)
	if err != nil {
		return fmt.Errorf("failed to insert security: %w", err)
	}

	// Insert into stocks table
	stocksQuery := `
		INSERT INTO stocks (ticker, cc, eps, teps, pe, tpe)
		VALUES (:ticker, :cc, :eps, :epsType, :pe, :peType)
	`
	_, err = tx.NamedExec(stocksQuery, stock)
	if err != nil {
		return fmt.Errorf("failed to insert stock: %w", err)
	}

	// Insert into dividends table if provided
	if stock.Security.Dividend != nil {
		dividendsQuery := `
			INSERT INTO dividends (
				ticker, cc, rate, trate, yield, tyield, ap, tap, pr, lgr, yog, lad, frequency, edd, pd
			) VALUES (
				:ticker, :cc, :rate, :rateType, :yield, :yieldType, :annualPayout, :apType,
				:payoutRatio, :growthRate, :yearsGrowth, :lastAnnounced, :frequency, :exDivDate, :payoutDate
			)
		`
		_, err = tx.NamedExec(dividendsQuery, stock.Security.Dividend)
		if err != nil {
			return fmt.Errorf("failed to insert dividend: %w", err)
		}
	}

	// Commit transaction
	return nil
}

func CreateETF(db *sqlx.DB, etf *ETF) error {
	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback()
			if err != nil {
				log.Errorf("Failed to rollback transaction: %v\n", err)
			}
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Errorf("Failed to rollback transaction: %v\n", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Errorf("Failed to commit transaction: %v\n", err)
			}
		}
	}()

	// Step 1: Insert into the securities table
	securitiesQuery := `
		INSERT INTO securities (
			ticker, cc, suffix, exchange, typology, fullname, price, pc, ppc,
			yrange, drange, marketcap, volume, avgvlm, beta, pclose, copen, bid, bidsz,
			ask, asksz, currency, created, updated
		) VALUES (
			:ticker, :cc, :suffix, :exchange, 'ETF', :fullname, :price, :pc, :ppc,
			:yrange, :drange, :marketcap, :volume, :avgvlm, :beta, :pclose, :copen, :bid, :bidsz,
			:ask, :asksz, :currency, :created, :updated
		)
	`
	_, err = tx.NamedExec(securitiesQuery, &etf.Security)
	if err != nil {
		return fmt.Errorf("failed to insert security: %w", err)
	}

	// Step 2: Insert into the etfs table
	etfsQuery := `
		INSERT INTO etfs (ticker, exchange, holdings, aum, er)
		VALUES (:ticker, :exchange, :holdings, :aum, :expenseRatio)
	`
	_, err = tx.NamedExec(etfsQuery, etf)
	if err != nil {
		return fmt.Errorf("failed to insert etf: %w", err)
	}

	// Step 3: Insert related securities into the etf_related_securities table
	if len(etf.RelatedSecurities) > 0 {
		relatedQuery := `
			INSERT INTO etf_related_securities (etf_ticker, etf_exchange, related_ticker, related_exchange)
			VALUES (:etf_ticker, :etf_exchange, :related_ticker, :related_exchange)
		`

		for _, related := range etf.RelatedSecurities {
			// Parse the related security string (e.g., "AAPL:NASDAQ")
			parts := strings.Split(related, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid related security format: %s", related)
			}

			_, err = tx.NamedExec(relatedQuery, map[string]interface{}{
				"etf_ticker":       etf.Ticker,
				"etf_exchange":     etf.Exchange,
				"related_ticker":   parts[0], // Ticker
				"related_exchange": parts[1], // Exchange
			})
			if err != nil {
				return fmt.Errorf("failed to insert related security '%s': %w", related, err)
			}
		}
	}

	// Step 4: Insert into dividends table if provided
	if etf.Security.Dividend != nil {
		dividendsQuery := `
			INSERT INTO dividends (
				ticker, exchange, rate, trate, yield, tyield, ap, tap, pr, lgr, yog, lad, frequency, edd, pd
			) VALUES (
				:ticker, :exchange, :rate, :trate, :yield, :tyield, :ap, :tap,
				:pr, :lgr, :yog, :lad, :frequency, :edd, :pd
			)
		`
		_, err = tx.NamedExec(dividendsQuery, etf.Security.Dividend)
		if err != nil {
			return fmt.Errorf("failed to insert dividend: %w", err)
		}
	}

	return nil
}

func CreateReit(db *sqlx.DB, reit *REIT) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback()
			if err != nil {
				log.Errorf("Failed to rollback transaction: %v\n", err)
			}
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Errorf("Failed to rollback transaction: %v\n", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Errorf("Failed to commit transaction: %v\n", err)
			}
		}
	}()

	// Insert into securities table
	securitiesQuery := `
		INSERT INTO securities (
			ticker, cc, suffix, exchange, typology, fullname, price, pc, ppc,
			yrange, drange, marketcap, volume, avgvlm, beta, pclose, copen, bid, bidsz,
			ask, asksz, currency, created, updated
		) VALUES (
			:ticker, :cc, :suffix, :exchange, 'reit', :fullname, :price, :pc, :ppc,
			:yrange, :drange, :marketcap, :volume, :avgvlm, :beta, :pclose, :copen, :bid, :bidsz,
			:ask, :asksz, :currency, :created, :updated
		)
	`
	_, err = tx.NamedExec(securitiesQuery, &reit.Security)
	if err != nil {
		return fmt.Errorf("failed to insert security: %w", err)
	}

	// Insert into reits table
	reitsQuery := `
		INSERT INTO reits (ticker, cc, ffo, tffo, pffo, tpffo)
		VALUES (:ticker, :cc, :ffo, :ffoType, :pffo, :pffoType)
	`
	_, err = tx.NamedExec(reitsQuery, reit)
	if err != nil {
		return fmt.Errorf("failed to insert reit: %w", err)
	}

	// Insert into dividends table if provided
	if reit.Security.Dividend != nil {
		dividendsQuery := `
			INSERT INTO dividends (
				ticker, cc, rate, trate, yield, tyield, ap, tap, pr, lgr, yog, lad, frequency, edd, pd
			) VALUES (
				:ticker, :cc, :rate, :rateType, :yield, :yieldType, :annualPayout, :apType,
				:payoutRatio, :growthRate, :yearsGrowth, :lastAnnounced, :frequency, :exDivDate, :payoutDate
			)
		`
		_, err = tx.NamedExec(dividendsQuery, reit.Security.Dividend)
		if err != nil {
			return fmt.Errorf("failed to insert dividend: %w", err)
		}
	}

	return nil
}

func GetStocks(
	db *sqlx.DB,
	exchange, country *string,
	minPrice, maxPrice *int,
	orderBy, orderDirection *string,
	limit *int,
	dividend bool,
) ([]Stock, error) {
	// Base query with dynamic JOIN clause
	query := `
		SELECT
			s.*,
			st.eps,
			st.teps AS epsType,
			st.pe,
			st.tpe AS peType,
			d.rate AS "dividend.rate",
			d.trate AS "dividend.rateType",
			d.yield AS "dividend.yield",
			d.tyield AS "dividend.yieldType",
			d.ap AS "dividend.annualPayout",
			d.tap AS "dividend.apType",
			d.pr AS "dividend.payoutRatio",
			d.lgr AS "dividend.growthRate",
			d.yog AS "dividend.yearsGrowth",
			d.lad AS "dividend.lastAnnounced",
			d.frequency AS "dividend.frequency",
			d.edd AS "dividend.exDivDate",
			d.pd AS "dividend.payoutDate"
		FROM securities s
		JOIN stocks st ON s.ticker = st.ticker AND s.exchange = st.exchange
	`

	// Adjust JOIN type based on dividend parameter
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.cc = d.cc"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.cc = d.cc"
	}

	// WHERE conditions
	query += `
		WHERE (:exchange IS NULL OR s.exchange = :exchange)
		  AND (:country IS NULL OR s.cc = :country)
		  AND (:minPrice IS NULL OR s.price >= :minPrice)
		  AND (:maxPrice IS NULL OR s.price <= :maxPrice)
	`

	// Ordering
	orderColumn := map[string]string{
		"price":     "s.price",
		"volume":    "s.volume",
		"avgvolume": "s.avgvlm",
		"marketcap": "s.marketcap",
		"pc":        "s.pc",
		"ppc":       "s.ppc",
		"updated":   "s.updated",
	}

	order := "s.price ASC"
	if orderBy != nil && *orderBy != "" {
		if col, exists := orderColumn[*orderBy]; exists {
			if orderDirection != nil && *orderDirection == "desc" {
				order = fmt.Sprintf("%s DESC", col)
			} else {
				order = fmt.Sprintf("%s ASC", col)
			}
		}
	}
	query += fmt.Sprintf(" ORDER BY %s", order)

	// Add limit
	if limit != nil && *limit > 0 {
		query += " LIMIT :limit"
	}

	// Query parameters
	params := map[string]interface{}{
		"exchange": exchange,
		"country":  country,
		"minPrice": minPrice,
		"maxPrice": maxPrice,
		"limit":    limit,
	}

	// Execute query
	var stocks []Stock
	err := db.Select(&stocks, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stocks: %w", err)
	}

	return stocks, nil
}

func GetStock(db *sqlx.DB, input string) (*Stock, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected '<ticker>:<exchange>'")
	}
	ticker, exchange := parts[0], parts[1]

	// Query to fetch stock details
	query := `
		SELECT
			s.*,
			st.eps,
			st.teps AS epsType,
			st.pe,
			st.tpe AS peType,
			d.rate AS "dividend.rate",
			d.trate AS "dividend.rateType",
			d.yield AS "dividend.yield",
			d.tyield AS "dividend.yieldType",
			d.ap AS "dividend.annualPayout",
			d.tap AS "dividend.apType",
			d.pr AS "dividend.payoutRatio",
			d.lgr AS "dividend.growthRate",
			d.yog AS "dividend.yearsGrowth",
			d.lad AS "dividend.lastAnnounced",
			d.frequency AS "dividend.frequency",
			d.edd AS "dividend.exDivDate",
			d.pd AS "dividend.payoutDate"
		FROM securities s
		JOIN stocks st ON s.ticker = st.ticker AND s.exchange = st.exchange
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange
	`

	// Execute the query
	var stock Stock
	err := db.Get(&stock, query, map[string]interface{}{
		"ticker":   ticker,
		"exchange": exchange,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("stock '%s' not found", input)
		}
		return nil, fmt.Errorf("failed to retrieve stock '%s': %w", input, err)
	}

	return &stock, nil
}

func GetETFs(
	db *sqlx.DB,
	exchange, country *string,
	minPrice, maxPrice *int,
	orderBy, orderDirection *string,
	limit *int,
	dividend bool,
) ([]ETF, error) {
	// Base query
	query := `
		SELECT
			s.*,
			e.holdings,
			e.aum,
			e.er AS expenseRatio,
			COALESCE(
				STRING_AGG(er.related_ticker || ':' || er.related_exchange, ','), ''
			) AS relatedSecurities,
			d.rate AS "dividend.rate",
			d.trate AS "dividend.rateType",
			d.yield AS "dividend.yield",
			d.tyield AS "dividend.yieldType",
			d.ap AS "dividend.annualPayout",
			d.tap AS "dividend.apType",
			d.pr AS "dividend.payoutRatio",
			d.lgr AS "dividend.growthRate",
			d.yog AS "dividend.yearsGrowth",
			d.lad AS "dividend.lastAnnounced",
			d.frequency AS "dividend.frequency",
			d.edd AS "dividend.exDivDate",
			d.pd AS "dividend.payoutDate"
		FROM securities s
		JOIN etfs e ON s.ticker = e.ticker AND s.exchange = e.exchange
		LEFT JOIN etf_related_securities er ON e.ticker = er.etf_ticker AND e.exchange = er.etf_exchange
	`

	// Adjust JOIN type for dividends based on the `dividend` parameter
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions
	query += `
		WHERE (:exchange IS NULL OR s.exchange = :exchange)
		  AND (:country IS NULL OR s.cc = :country)
		  AND (:minPrice IS NULL OR s.price >= :minPrice)
		  AND (:maxPrice IS NULL OR s.price <= :maxPrice)
	`

	// Apply ordering
	orderColumn := map[string]string{
		"price":     "s.price",
		"volume":    "s.volume",
		"avgvolume": "s.avgvlm",
		"marketcap": "s.marketcap",
		"pc":        "s.pc",
		"ppc":       "s.ppc",
		"updated":   "s.updated",
	}

	order := "s.price ASC" // Default ordering
	if orderBy != nil && *orderBy != "" {
		if col, exists := orderColumn[*orderBy]; exists {
			if orderDirection != nil && *orderDirection == "desc" {
				order = fmt.Sprintf("%s DESC", col)
			} else {
				order = fmt.Sprintf("%s ASC", col)
			}
		}
	}
	query += fmt.Sprintf(" ORDER BY %s", order)

	// Add limit
	if limit != nil && *limit > 0 {
		query += " LIMIT :limit"
	}

	// Query parameters
	params := map[string]interface{}{
		"exchange": exchange,
		"country":  country,
		"minPrice": minPrice,
		"maxPrice": maxPrice,
		"limit":    limit,
	}

	// Execute the query
	rows, err := db.Queryx(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ETFs: %w", err)
	}
	defer rows.Close()

	// Parse results
	var etfs []ETF
	for rows.Next() {
		var etf ETF
		if err := etf.Scan(rows); err != nil {
			return nil, fmt.Errorf("failed to scan ETF row: %w", err)
		}
		etfs = append(etfs, etf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over ETF rows: %w", err)
	}

	return etfs, nil
}

func GetETF(db *sqlx.DB, input string) (*ETF, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected '<ticker>:<exchange>'")
	}
	ticker, exchange := parts[0], parts[1]

	// Query to fetch ETF details
	query := `
		SELECT
			s.*,
			e.holdings,
			e.aum,
			e.er AS expenseRatio,
			COALESCE(
				STRING_AGG(er.related_ticker || ':' || er.related_exchange, ','), ''
			) AS relatedSecurities,
			d.rate AS "dividend.rate",
			d.trate AS "dividend.rateType",
			d.yield AS "dividend.yield",
			d.tyield AS "dividend.yieldType",
			d.ap AS "dividend.annualPayout",
			d.tap AS "dividend.apType",
			d.pr AS "dividend.payoutRatio",
			d.lgr AS "dividend.growthRate",
			d.yog AS "dividend.yearsGrowth",
			d.lad AS "dividend.lastAnnounced",
			d.frequency AS "dividend.frequency",
			d.edd AS "dividend.exDivDate",
			d.pd AS "dividend.payoutDate"
		FROM securities s
		JOIN etfs e ON s.ticker = e.ticker AND s.exchange = e.exchange
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		LEFT JOIN etf_related_securities er ON e.ticker = er.etf_ticker AND e.exchange = er.etf_exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange
		GROUP BY s.ticker, s.exchange, e.holdings, e.aum, e.er, d.rate, d.trate, d.yield, d.tyield,
		         d.ap, d.tap, d.pr, d.lgr, d.yog, d.lad, d.frequency, d.edd, d.pd
	`

	// Execute the query
	var etf ETF
	err := db.Get(&etf, query, map[string]interface{}{
		"ticker":   ticker,
		"exchange": exchange,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ETF '%s' not found", input)
		}
		return nil, fmt.Errorf("failed to retrieve ETF '%s': %w", input, err)
	}

	// Parse related securities
	if etf.RelatedSecurities != nil {
		etf.RelatedSecurities = strings.Split(etf.RelatedSecurities[0], ",")
	}

	return &etf, nil
}

func GetREITs(
	db *sqlx.DB,
	exchange, country *string,
	minPrice, maxPrice *int,
	orderBy, orderDirection *string,
	limit *int,
	dividend bool,
) ([]REIT, error) {
	// Base query
	query := `
		SELECT
			s.*,
			r.ffo,
			r.tffo AS ffoType,
			r.pffo,
			r.tpffo AS pffoType,
			d.rate AS "dividend.rate",
			d.trate AS "dividend.rateType",
			d.yield AS "dividend.yield",
			d.tyield AS "dividend.yieldType",
			d.ap AS "dividend.annualPayout",
			d.tap AS "dividend.apType",
			d.pr AS "dividend.payoutRatio",
			d.lgr AS "dividend.growthRate",
			d.yog AS "dividend.yearsGrowth",
			d.lad AS "dividend.lastAnnounced",
			d.frequency AS "dividend.frequency",
			d.edd AS "dividend.exDivDate",
			d.pd AS "dividend.payoutDate"
		FROM securities s
		JOIN reits r ON s.ticker = r.ticker AND s.exchange = r.exchange
	`

	// Adjust JOIN type for dividends based on the `dividend` parameter
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions
	query += `
		WHERE (:exchange IS NULL OR s.exchange = :exchange)
		  AND (:country IS NULL OR s.cc = :country)
		  AND (:minPrice IS NULL OR s.price >= :minPrice)
		  AND (:maxPrice IS NULL OR s.price <= :maxPrice)
	`

	// Apply ordering
	orderColumn := map[string]string{
		"price":     "s.price",
		"volume":    "s.volume",
		"avgvolume": "s.avgvlm",
		"marketcap": "s.marketcap",
		"pc":        "s.pc",
		"ppc":       "s.ppc",
		"updated":   "s.updated",
	}

	order := "s.price ASC" // Default ordering
	if orderBy != nil && *orderBy != "" {
		if col, exists := orderColumn[*orderBy]; exists {
			if orderDirection != nil && *orderDirection == "desc" {
				order = fmt.Sprintf("%s DESC", col)
			} else {
				order = fmt.Sprintf("%s ASC", col)
			}
		}
	}
	query += fmt.Sprintf(" ORDER BY %s", order)

	// Add limit
	if limit != nil && *limit > 0 {
		query += " LIMIT :limit"
	}

	// Query parameters
	params := map[string]interface{}{
		"exchange": exchange,
		"country":  country,
		"minPrice": minPrice,
		"maxPrice": maxPrice,
		"limit":    limit,
	}

	// Execute the query
	rows, err := db.Queryx(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve REITs: %w", err)
	}
	defer rows.Close()

	// Parse results
	var reits []REIT
	for rows.Next() {
		var reit REIT
		if err := reit.Scan(rows); err != nil {
			return nil, fmt.Errorf("failed to scan REIT row: %w", err)
		}
		reits = append(reits, reit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over REIT rows: %w", err)
	}

	return reits, nil
}
