package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/jmoiron/sqlx"
)

type Timing string
type Frequency string

const (
	TimingFWD Timing = "fwd"
	TimingTTM Timing = "ttm"

	FrequencyUnknown   Frequency = "unknown"
	FrequencyWeekly    Frequency = "weekly"
	FrequencyBiweekly  Frequency = "biweekly"
	FrequencyMonthly   Frequency = "monthly"
	FrequencyQuarterly Frequency = "quarterly"
	FrequencySemi      Frequency = "semi"
	FrequencyYearly    Frequency = "yearly"
)

type Country struct {
	Code      string `db:"code" json:"code"`
	Label     string `db:"label" json:"label"`
	Currency  string `db:"currency" json:"currency"`
	Continent string `db:"continent,omitempty" json:"continent,omitempty"`
	ISO       string `db:"iso,omitempty" json:"iso,omitempty"`
}

type Exchange struct {
	Title     string         `db:"title" json:"title"`
	Prefix    sql.NullString `db:"prefix" json:"prefix,omitempty"`
	Suffix    sql.NullString `db:"suffix" json:"suffix,omitempty"`
	CC        string         `db:"cc" json:"countryCode"`
	OpenTime  sql.NullTime   `db:"opentime" json:"openTime,omitempty"`
	CloseTime sql.NullTime   `db:"closetime" json:"closeTime,omitempty"`
}

type Security struct {
	Ticker      string         `db:"ticker" json:"ticker"`
	Exchange    string         `db:"exchange" json:"exchange"`
	Typology    string         `db:"typology" json:"typology"` // STOCK, ETF, REIT
	FullName    string         `db:"fullname" json:"fullName"`
	Sector      sql.NullString `db:"sector" json:"sector,omitempty"`
	Industry    sql.NullString `db:"industry" json:"industry,omitempty"`
	SubIndustry sql.NullString `db:"subindustry" json:"subIndustry,omitempty"`
	Price       int            `db:"price" json:"price"`
	PC          int            `db:"pc" json:"pc"`
	PCP         int            `db:"PCP" json:"PCP"`
	YearLow     int            `db:"yrl" json:"yearLow"`
	YearHigh    int            `db:"yrh" json:"yearHigh"`
	DayLow      int            `db:"drl" json:"dayLow"`
	DayHigh     int            `db:"drh" json:"dayHigh"`
	Consensus   sql.NullString `db:"consensus" json:"consensus,omitempty"`
	Score       sql.NullInt64  `db:"score" json:"score,omitempty"`
	Coverage    sql.NullInt64  `db:"coverage" json:"coverage,omitempty"`
	MarketCap   sql.NullInt64  `db:"cap" json:"marketCap,omitempty"`
	Volume      sql.NullInt64  `db:"volume" json:"volume,omitempty"`
	AvgVolume   sql.NullInt64  `db:"avgvolume" json:"avgVolume,omitempty"`
	Outstanding sql.NullInt64  `db:"outstanding" json:"outstanding,omitempty"`
	Beta        sql.NullInt64  `db:"beta" json:"beta,omitempty"`
	PClose      int            `db:"pclose" json:"previousClose"`
	COpen       int            `db:"copen" json:"currentOpen"`
	Bid         int            `db:"bid" json:"bid"`
	BidSize     sql.NullInt64  `db:"bidsz" json:"bidSize,omitempty"`
	Ask         int            `db:"ask" json:"ask"`
	AskSize     sql.NullInt64  `db:"asksz" json:"askSize,omitempty"`
	EPS         sql.NullInt64  `db:"eps" json:"eps,omitempty"`
	PE          sql.NullInt64  `db:"pe" json:"pe,omitempty"`
	STM         sql.NullString `db:"stm" json:"stm,omitempty"` // Enum: fwd, ttm
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
		s.Exchange = data["exchange"].(string)
		s.Typology = data["typology"].(string)
		s.FullName = data["fullname"].(string)
		s.Price = int(data["price"].(int64))
		s.PC = int(data["pc"].(int64))
		s.PCP = int(data["ppc"].(int64))
		s.YearLow = int(data["yrl"].(int64))
		s.YearHigh = int(data["yrh"].(int64))
		s.DayLow = int(data["drl"].(int64))
		s.DayHigh = int(data["drh"].(int64))
		s.PClose = int(data["pclose"].(int64))
		s.COpen = int(data["copen"].(int64))
		s.Bid = int(data["bid"].(int64))
		s.Ask = int(data["ask"].(int64))

		// Handle nullable fields
		if sector, ok := data["sector"].(string); ok {
			s.Sector = sql.NullString{String: sector, Valid: true}
		}
		if industry, ok := data["industry"].(string); ok {
			s.Industry = sql.NullString{String: industry, Valid: true}
		}
		if subIndustry, ok := data["subindustry"].(string); ok {
			s.SubIndustry = sql.NullString{String: subIndustry, Valid: true}
		}
		if consensus, ok := data["consensus"].(string); ok {
			s.Consensus = sql.NullString{String: consensus, Valid: true}
		}
		if score, ok := data["score"].(int64); ok {
			s.Score = sql.NullInt64{Int64: score, Valid: true}
		}
		if coverage, ok := data["coverage"].(int64); ok {
			s.Coverage = sql.NullInt64{Int64: coverage, Valid: true}
		}
		if marketcap, ok := data["marketcap"].(int64); ok {
			s.MarketCap = sql.NullInt64{Int64: int64(marketcap), Valid: true}
		}
		if volume, ok := data["volume"].(int64); ok {
			s.Volume = sql.NullInt64{Int64: volume, Valid: true}
		}
		if avgVolume, ok := data["avgvolume"].(int64); ok {
			s.AvgVolume = sql.NullInt64{Int64: avgVolume, Valid: true}
		}
		if outstanding, ok := data["outstanding"].(int64); ok {
			s.Outstanding = sql.NullInt64{Int64: outstanding, Valid: true}
		}
		if beta, ok := data["beta"].(int64); ok {
			s.Beta = sql.NullInt64{Int64: beta, Valid: true}
		}
		if eps, ok := data["eps"].(int64); ok {
			s.EPS = sql.NullInt64{Int64: eps, Valid: true}
		}
		if pe, ok := data["pe"].(int64); ok {
			s.PE = sql.NullInt64{Int64: pe, Valid: true}
		}
		if stm, ok := data["stm"].(string); ok {
			s.STM = sql.NullString{String: stm, Valid: true}
		}
		if bidSize, ok := data["bidsz"].(int64); ok {
			s.BidSize = sql.NullInt64{Int64: bidSize, Valid: true}
		}
		if askSize, ok := data["asksz"].(int64); ok {
			s.AskSize = sql.NullInt64{Int64: askSize, Valid: true}
		}

		// Handle related Dividend data
		if dividendData, ok := data["dividend"].(map[string]any); ok {
			dividend := &Dividend{}
			dividend.Ticker = dividendData["ticker"].(string)
			dividend.Exchange = dividendData["exchange"].(string)
			dividend.Yield = dividendData["yield"].(int)

			//Hanle nullable fields
			if annualPayout, ok := dividendData["ap"].(int64); ok {
				dividend.AnnualPayout = sql.NullInt64{Int64: annualPayout, Valid: true}
			}
			if timing, ok := dividendData["tm"].(string); ok {
				dividend.Timing = sql.NullString{String: timing, Valid: true}
			}
			if payoutRatio, ok := dividendData["pr"].(int64); ok {
				dividend.PayoutRatio = sql.NullInt64{Int64: payoutRatio, Valid: true}
			}
			if growthRate, ok := dividendData["lgr"].(int64); ok {
				dividend.GrowthRate = sql.NullInt64{Int64: growthRate, Valid: true}
			}
			if yearsGrowth, ok := dividendData["yog"].(int64); ok {
				dividend.YearsGrowth = sql.NullInt64{Int64: yearsGrowth, Valid: true}
			}
			if frequency, ok := dividendData["frequency"].(string); ok {
				dividend.Frequency = sql.NullString{String: frequency, Valid: true}
			}
			if lastAnnounced, ok := dividendData["lad"].(int64); ok {
				dividend.LastAnnounced = sql.NullInt64{Int64: lastAnnounced, Valid: true}
			}
			if exDate, ok := dividendData["edd"].(time.Time); ok {
				dividend.ExDivDate = sql.NullTime{Time: exDate, Valid: true}
			}
			if payDate, ok := dividendData["pd"].(time.Time); ok {
				dividend.PayoutDate = sql.NullTime{Time: payDate, Valid: true}
			}
			s.Dividend = dividend
		}

	default:
		return fmt.Errorf("unsupported Scan source: %T", src)
	}

	return nil
}

func InsertSecurity(tx *sqlx.Tx, security *Security) error {
	query := `
		INSERT INTO securities (
			ticker, exchange, typology, fullname, sector, industry, subindustry, price, pc, PCP,
			yrl, yrh, drl, drh, consensus, score, coverage, cap, volume, avgvolume, outstanding,
			beta, pclose, copen, bid, bidsz, ask, asksz, eps, pe, stm
		) VALUES (
			:ticker, :exchange, :typology, :fullname, :sector, :industry, :subindustry, :price, :pc, :PCP,
			:yrl, :yrh, :drl, :drh, :consensus, :score, :coverage, :cap, :volume, :avgvolume, :outstanding,
			:beta, :pclose, :copen, :bid, :bidsz, :ask, :asksz, :eps, :pe, :stm
		)
	`

	_, err := tx.NamedExec(query, security)
	if err != nil {
		return fmt.Errorf("failed to insert security (%s): %w", security.Ticker, err)
	}

	return nil
}

// Dividend represents a row from the dividends table.
type Dividend struct {
	Ticker        string         `db:"ticker" json:"ticker"`
	Exchange      string         `db:"exchange" json:"exchange"`
	Yield         int            `db:"yield" json:"yield"`
	AnnualPayout  sql.NullInt64  `db:"ap" json:"annualPayout,omitempty"`
	Timing        sql.NullString `db:"tm" json:"timing,omitempty"` // Enum: fwd, ttm
	PayoutRatio   sql.NullInt64  `db:"pr" json:"payoutRatio,omitempty"`
	GrowthRate    sql.NullInt64  `db:"lgr" json:"growthRate,omitempty"`
	YearsGrowth   sql.NullInt64  `db:"yog" json:"yearsGrowth,omitempty"`
	LastAnnounced sql.NullInt64  `db:"lad" json:"lastAnnounced,omitempty"`
	Frequency     sql.NullString `db:"frequency" json:"frequency,omitempty"` // Enum: Frequency
	ExDivDate     sql.NullTime   `db:"edd" json:"exDivDate,omitempty"`
	PayoutDate    sql.NullTime   `db:"pd" json:"payoutDate,omitempty"`
}

func InsertDividend(tx *sqlx.Tx, dividend *Dividend) error {
	if dividend == nil {
		return nil // No dividend, skip insertion
	}

	query := `
		INSERT INTO dividends (
			ticker, exchange, yield, ap, tm, pr, lgr, yog, lad, frequency, edd, pd
		) VALUES (
			:ticker, :exchange, :yield, :annualPayout, :timing, :payoutRatio,
			:growthRate, :yearsGrowth, :lastAnnounced, :frequency, :exDivDate, :payoutDate
		)
	`

	_, err := tx.NamedExec(query, dividend)
	if err != nil {
		return fmt.Errorf("failed to insert dividend for %s (%s): %w", dividend.Ticker, dividend.Exchange, err)
	}

	return nil
}

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

// ETF represents a row from the etfs table.
type ETF struct {
	Security          `json:"security"` // Embedded security properties
	Holdings          int               `db:"holdings" json:"holdings"`
	AUM               sql.NullInt64     `db:"aum" json:"aum,omitempty"`
	ExpenseRatio      sql.NullInt64     `db:"er" json:"expenseRatio,omitempty"`
	NAV               sql.NullInt64     `db:"nav" json:"nav,omitempty"`
	InceptionDate     sql.NullTime      `db:"inception" json:"inception,omitempty"`
	RelatedSecurities []string          `json:"relatedSecurities"` // Related securities as "TICKER:EXCHANGE:ALLOCATION"
}

func (etf *ETF) flatten() map[string]interface{} {
	return map[string]interface{}{
		"ticker":       etf.Security.Ticker,
		"exchange":     etf.Security.Exchange,
		"holdings":     etf.Holdings,
		"aum":          etf.AUM,
		"expenseRatio": etf.ExpenseRatio,
		"nav":          etf.NAV,
		"inception":    etf.InceptionDate,
	}
}

func (etf *ETF) Scan(rows *sqlx.Rows) error {
	// Temporary variables for related securities and nullable fields
	var relatedSecurities string

	// Scan all fields
	err := rows.Scan(
		&etf.Security.Ticker,
		&etf.Security.Exchange,
		&etf.Security.Typology,
		&etf.Security.FullName,
		&etf.Security.Sector,
		&etf.Security.Industry,
		&etf.Security.SubIndustry,
		&etf.Security.Price,
		&etf.Security.PC,
		&etf.Security.PCP,
		&etf.Security.YearHigh,
		&etf.Security.YearLow,
		&etf.Security.DayHigh,
		&etf.Security.DayLow,
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
		&etf.Security.Created,
		&etf.Security.Updated,
		&etf.Holdings,
		&etf.AUM,
		&etf.ExpenseRatio,
		&relatedSecurities, // Comma-separated related securities
		&etf.Security.Dividend.Yield,
		&etf.Security.Dividend.Timing,
		&etf.Security.Dividend.AnnualPayout,
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
		INSERT INTO etfs (ticker, exchange, holdings, aum, er, nav, inception)
		VALUES (:ticker, :exchange, :holdings, :aum, :expenseRatio, :nav, :inception)
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

			_, err = tx.NamedExec(relatedQuery, map[string]interface{}{
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

// REIT represents a row from the reits table.
type REIT struct {
	Security   `json:"security"` // Embedded security properties
	Occupation sql.NullInt64     `db:"occupation" json:"occupation,omitempty"`
	Focus      sql.NullString    `db:"focus" json:"focus,omitempty"`
	FFO        sql.NullInt64     `db:"ffo" json:"ffo,omitempty"`
	PFFO       sql.NullInt64     `db:"pffo" json:"pffo,omitempty"`
	Timing     sql.NullString    `db:"tm" json:"timing,omitempty"` // Enum: fwd, ttm
}

func (reit *REIT) flatten() map[string]interface{} {
	return map[string]interface{}{
		"ticker":     reit.Security.Ticker,
		"occupation": reit.Security.Exchange,
		"focus":      reit.Focus,
		"ffo":        reit.FFO,
		"pffo":       reit.PFFO,
		"tm":         reit.Timing,
	}
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

func GetStock(db *sqlx.DB, input string) (*Security, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected '<ticker>:<exchange>'")
	}
	ticker, exchange := parts[0], parts[1]

	// Query for retrieving stock details, including dividend (if available)
	query := `
		SELECT
			s.*,
			d.yield, d.tm AS timing, d.ap AS annualPayout, d.pr AS payoutRatio,
			d.lgr AS growthRate, d.yog AS yearsGrowth, d.lad AS lastAnnounced,
			d.frequency, d.edd AS exDivDate, d.pd AS payoutDate
		FROM securities s
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange AND s.typology = 'STOCK'
	`

	// Execute the query
	var stock Security
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

func GetStocks(
	db *sqlx.DB,
	exchange, country *string,
	minPrice, maxPrice *int,
	orderBy, orderDirection *string,
	limit *int,
	dividend bool,
) ([]Security, error) {
	// Base query selecting relevant security fields where typology = 'STOCK'
	query := `
		SELECT
			s.ticker, s.exchange, s.typology, s.fullname, s.sector, s.industry, s.subindustry,
			s.price, s.pc, s.pcc, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
			s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
			s.ask, s.asksz, s.eps, s.pe, s.stm, s.created, s.updated,
			d.yield, d.tm AS timing, d.ap AS annualPayout, d.pr AS payoutRatio,
			d.lgr AS growthRate, d.yog AS yearsGrowth, d.lad AS lastAnnounced,
			d.frequency, d.edd AS exDivDate, d.pd AS payoutDate
		FROM securities s
		WHERE s.typology = 'STOCK'
	`

	// Adjust JOIN type based on dividend presence
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions
	query += `
		AND (:exchange IS NULL OR s.exchange = :exchange)
		AND (:country IS NULL OR s.exchange IN (SELECT title FROM exchanges WHERE cc = :country))
		AND (:minPrice IS NULL OR s.price >= :minPrice)
		AND (:maxPrice IS NULL OR s.price <= :maxPrice)
	`

	// Apply ordering
	orderColumn := map[string]string{
		"price":     "s.price",
		"volume":    "s.volume",
		"avgvolume": "s.avgvolume",
		"marketcap": "s.cap",
		"pc":        "s.pc",
		"pcc":       "s.pcc",
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

func GetETF(db *sqlx.DB, input string) (*Security, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected '<ticker>:<exchange>'")
	}
	ticker, exchange := parts[0], parts[1]

	// Query for retrieving ETF details, including dividend (if available)
	query := `
		SELECT
			s.*,
			d.yield, d.tm AS timing, d.ap AS annualPayout, d.pr AS payoutRatio,
			d.lgr AS growthRate, d.yog AS yearsGrowth, d.lad AS lastAnnounced,
			d.frequency, d.edd AS exDivDate, d.pd AS payoutDate
		FROM securities s
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange AND s.typology = 'ETF'
	`

	// Execute the query
	var etf Security
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

	return &etf, nil
}

func GetETFs(
	db *sqlx.DB,
	exchange, country *string,
	minPrice, maxPrice *int,
	orderBy, orderDirection *string,
	limit *int,
	dividend bool,
) ([]Security, error) {
	// Base query selecting relevant security fields where typology = 'ETF'
	query := `
		SELECT
			s.ticker, s.exchange, s.typology, s.fullname, s.sector, s.industry, s.subindustry,
			s.price, s.pc, s.pcc, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
			s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
			s.ask, s.asksz, s.eps, s.pe, s.stm, s.created, s.updated,
			d.yield, d.tm AS timing, d.ap AS annualPayout, d.pr AS payoutRatio,
			d.lgr AS growthRate, d.yog AS yearsGrowth, d.lad AS lastAnnounced,
			d.frequency, d.edd AS exDivDate, d.pd AS payoutDate
		FROM securities s
		WHERE s.typology = 'ETF'
	`

	// Adjust JOIN type based on dividend presence
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions
	query += `
		AND (:exchange IS NULL OR s.exchange = :exchange)
		AND (:country IS NULL OR s.exchange IN (SELECT title FROM exchanges WHERE cc = :country))
		AND (:minPrice IS NULL OR s.price >= :minPrice)
		AND (:maxPrice IS NULL OR s.price <= :maxPrice)
	`

	// Apply ordering
	orderColumn := map[string]string{
		"price":     "s.price",
		"volume":    "s.volume",
		"avgvolume": "s.avgvolume",
		"marketcap": "s.cap",
		"pc":        "s.pc",
		"pcc":       "s.pcc",
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

	// Execute the query
	rows, err := db.Queryx(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ETFs: %w", err)
	}
	defer rows.Close()

	// Parse results
	var etfs []Security
	for rows.Next() {
		var etf Security
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

func GetREIT(db *sqlx.DB, input string) (*Security, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected '<ticker>:<exchange>'")
	}
	ticker, exchange := parts[0], parts[1]

	// Query for retrieving REIT details, including dividend (if available)
	query := `
		SELECT
			s.*,
			d.yield, d.tm AS timing, d.ap AS annualPayout, d.pr AS payoutRatio,
			d.lgr AS growthRate, d.yog AS yearsGrowth, d.lad AS lastAnnounced,
			d.frequency, d.edd AS exDivDate, d.pd AS payoutDate
		FROM securities s
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange AND s.typology = 'REIT'
	`

	// Execute the query
	var reit Security
	err := db.Get(&reit, query, map[string]interface{}{
		"ticker":   ticker,
		"exchange": exchange,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("REIT '%s' not found", input)
		}
		return nil, fmt.Errorf("failed to retrieve REIT '%s': %w", input, err)
	}

	return &reit, nil
}

func GetREITs(
	db *sqlx.DB,
	exchange, country *string,
	minPrice, maxPrice *int,
	orderBy, orderDirection *string,
	limit *int,
	dividend bool,
) ([]Security, error) {
	// Base query selecting relevant security fields where typology = 'REIT'
	query := `
		SELECT
			s.ticker, s.exchange, s.typology, s.fullname, s.sector, s.industry, s.subindustry,
			s.price, s.pc, s.pcc, s.yrl, s.yrh, s.drl, s.drh, s.consensus, s.score, s.coverage,
			s.cap, s.volume, s.avgvolume, s.outstanding, s.beta, s.pclose, s.copen, s.bid, s.bidsz,
			s.ask, s.asksz, s.eps, s.pe, s.stm, s.created, s.updated,
			d.yield, d.tm AS timing, d.ap AS annualPayout, d.pr AS payoutRatio,
			d.lgr AS growthRate, d.yog AS yearsGrowth, d.lad AS lastAnnounced,
			d.frequency, d.edd AS exDivDate, d.pd AS payoutDate
		FROM securities s
		WHERE s.typology = 'REIT'
	`

	// Adjust JOIN type based on dividend presence
	if dividend {
		query += " INNER JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	} else {
		query += " LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange"
	}

	// WHERE conditions
	query += `
		AND (:exchange IS NULL OR s.exchange = :exchange)
		AND (:country IS NULL OR s.exchange IN (SELECT title FROM exchanges WHERE cc = :country))
		AND (:minPrice IS NULL OR s.price >= :minPrice)
		AND (:maxPrice IS NULL OR s.price <= :maxPrice)
	`

	// Apply ordering
	orderColumn := map[string]string{
		"price":     "s.price",
		"volume":    "s.volume",
		"avgvolume": "s.avgvolume",
		"marketcap": "s.cap",
		"pc":        "s.pc",
		"pcc":       "s.pcc",
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

	// Execute the query
	rows, err := db.Queryx(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve REITs: %w", err)
	}
	defer rows.Close()

	// Parse results
	var reits []Security
	for rows.Next() {
		var reit Security
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
