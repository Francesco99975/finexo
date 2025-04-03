package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Security struct {
	Ticker      string         `db:"ticker" json:"ticker"`
	Exchange    string         `db:"exchange" json:"exchange"`
	Typology    string         `db:"typology" json:"typology"` // STOCK, ETF, REIT
	Currency    string         `db:"currency" json:"currency"`
	FullName    string         `db:"fullname" json:"fullName"`
	Sector      NullableString `db:"sector" json:"sector,omitempty"`
	Industry    NullableString `db:"industry" json:"industry,omitempty"`
	SubIndustry NullableString `db:"subindustry" json:"subIndustry,omitempty"`
	Price       int            `db:"price" json:"price"`
	PC          int            `db:"pc" json:"pc"`
	PCP         int            `db:"pcp" json:"pcp"`
	YearLow     int            `db:"yrl" json:"yearLow"`
	YearHigh    int            `db:"yrh" json:"yearHigh"`
	DayLow      int            `db:"drl" json:"dayLow"`
	DayHigh     int            `db:"drh" json:"dayHigh"`
	Consensus   NullableString `db:"consensus" json:"consensus,omitempty"`
	Score       NullableInt    `db:"score" json:"score,omitempty"`
	Coverage    NullableInt    `db:"coverage" json:"coverage,omitempty"`
	MarketCap   NullableInt    `db:"cap" json:"marketCap,omitempty"`
	Volume      NullableInt    `db:"volume" json:"volume,omitempty"`
	AvgVolume   NullableInt    `db:"avgvolume" json:"avgVolume,omitempty"`
	Outstanding NullableInt    `db:"outstanding" json:"outstanding,omitempty"`
	Beta        NullableInt    `db:"beta" json:"beta,omitempty"`
	PClose      int            `db:"pclose" json:"previousClose"`
	COpen       int            `db:"copen" json:"currentOpen"`
	Bid         int            `db:"bid" json:"bid"`
	BidSize     NullableInt    `db:"bidsz" json:"bidSize,omitempty"`
	Ask         int            `db:"ask" json:"ask"`
	AskSize     NullableInt    `db:"asksz" json:"askSize,omitempty"`
	EPS         NullableInt    `db:"eps" json:"eps,omitempty"`
	PE          NullableInt    `db:"pe" json:"pe,omitempty"`
	Target      NullableInt    `db:"target" json:"target,omitempty"`
	STM         NullableString `db:"stm" json:"stm,omitempty"` // Enum: fwd, ttm/
	Created     time.Time      `db:"created" json:"created"`
	Updated     time.Time      `db:"updated" json:"updated"`

	Dividend *Dividend `db:"-" json:"dividend,omitempty"` // Associated dividend data (if exists)
}

// Scan implements the sql.Scanner interface for Security.
func (s *Security) Scan(rows *sqlx.Rows) error {
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
		&s.Ticker, &s.Exchange, &s.Typology, &s.Currency, &s.FullName, &s.Sector, &s.Industry, &s.SubIndustry,
		&s.Price, &s.PC, &s.PCP, &s.YearLow, &s.YearHigh, &s.DayLow, &s.DayHigh, &s.Consensus, &s.Score, &s.Coverage,
		&s.MarketCap, &s.Volume, &s.AvgVolume, &s.Outstanding, &s.Beta,
		&s.PClose, &s.COpen, &s.Bid, &s.BidSize, &s.Ask, &s.AskSize,
		&s.EPS, &s.PE, &s.Target, &s.STM, &s.Created, &s.Updated,

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
		s.Dividend = &Dividend{
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

func (s *Security) CreatePrettyPrintString() string {
	var sb strings.Builder
	sb.WriteString("Ticker: " + s.Ticker + " -- ")
	sb.WriteString("Exchange: " + s.Exchange + " -- ")
	sb.WriteString("Typology: " + s.Typology + " -- ")
	sb.WriteString("Currency: " + s.Currency + " -- ")
	sb.WriteString("Full Name: " + s.FullName + " -- ")
	sb.WriteString("Sector: " + s.Sector.String + " -- ")
	sb.WriteString("Industry: " + s.Industry.String + " -- ")
	sb.WriteString("SubIndustry: " + s.SubIndustry.String + " -- ")
	sb.WriteString("Consensus: " + s.Consensus.String + " -- ")
	if s.Score.Valid {
		sb.WriteString("Score: " + strconv.Itoa(int(s.Score.Int64)) + " -- ")
	}
	if s.Coverage.Valid {
		sb.WriteString("Coverage: " + strconv.Itoa(int(s.Coverage.Int64)) + " -- ")
	}
	if s.MarketCap.Valid {
		sb.WriteString("Market Cap: " + strconv.Itoa(int(s.MarketCap.Int64)) + " -- ")
	}
	if s.Volume.Valid {
		sb.WriteString("Volume: " + strconv.Itoa(int(s.Volume.Int64)) + " -- ")
	}
	if s.AvgVolume.Valid {
		sb.WriteString("Avg Volume: " + strconv.Itoa(int(s.AvgVolume.Int64)) + " -- ")
	}
	if s.Outstanding.Valid {
		sb.WriteString("Outstanding: " + strconv.Itoa(int(s.Outstanding.Int64)) + " -- ")
	}
	if s.Beta.Valid {
		sb.WriteString("Beta: " + strconv.Itoa(int(s.Beta.Int64)) + " -- ")
	}
	if s.EPS.Valid {
		sb.WriteString("EPS: " + strconv.Itoa(int(s.EPS.Int64)) + " -- ")
	}
	if s.PE.Valid {
		sb.WriteString("PE: " + strconv.Itoa(int(s.PE.Int64)) + " -- ")
	}
	if s.Target.Valid {
		sb.WriteString("Target: " + strconv.Itoa(int(s.Target.Int64)) + " -- ")
	}
	if s.STM.Valid {
		sb.WriteString("STM: " + s.STM.String + " -- ")
	}
	if s.BidSize.Valid {
		sb.WriteString("Bid Size: " + strconv.Itoa(int(s.BidSize.Int64)) + " -- ")
	}
	if s.AskSize.Valid {
		sb.WriteString("Ask Size: " + strconv.Itoa(int(s.AskSize.Int64)) + " -- ")
	}

	if s.Dividend != nil {
		sb.WriteString("Dividend --> ")
		sb.WriteString(s.Dividend.PrettyPrintString())
	}

	return sb.String()
}

func IsSecuritiesTableEmpty(db *sqlx.DB) (bool, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM securities")
	if err != nil {
		return false, fmt.Errorf("failed to check securities table: %w", err)
	}

	return count == 0, nil
}

func InsertSecurity(tx *sqlx.Tx, security *Security) error {
	query := `
		INSERT INTO securities (
			ticker, exchange, typology, currency, fullname, sector, industry, subindustry, price, pc, pcp,
			yrl, yrh, drl, drh, consensus, score, coverage, cap, volume, avgvolume, outstanding,
			beta, pclose, copen, bid, bidsz, ask, asksz, eps, pe, target, stm
		) VALUES (
			:ticker, :exchange, :typology, :currency, :fullname, :sector, :industry, :subindustry, :price, :pc, :pcp,
			:yrl, :yrh, :drl, :drh, :consensus, :score, :coverage, :cap, :volume, :avgvolume, :outstanding,
			:beta, :pclose, :copen, :bid, :bidsz, :ask, :asksz, :eps, :pe, :target, :stm
		)
	`

	_, err := tx.NamedExec(query, security)
	if err != nil {
		return fmt.Errorf("failed to insert security (%s): %w", security.Ticker, err)
	}

	return nil
}

func UpdateSecurity(tx *sqlx.Tx, security *Security) error {
	if security.Ticker == "" {
		return fmt.Errorf("ticker is required for update")
	}

	query := "UPDATE securities SET "
	args := make(map[string]any)
	updates := []string{}

	// String fields
	if security.Currency != "" {
		updates = append(updates, "currency = :currency")
		args["currency"] = security.Currency
	}
	if security.FullName != "" {
		updates = append(updates, "fullname = :fullname")
		args["fullname"] = security.FullName
	}

	// Nullable String Fields
	if security.Sector.Valid {
		updates = append(updates, "sector = :sector")
		args["sector"] = security.Sector.String
	}
	if security.Industry.Valid {
		updates = append(updates, "industry = :industry")
		args["industry"] = security.Industry.String
	}
	if security.SubIndustry.Valid {
		updates = append(updates, "subindustry = :subindustry")
		args["subindustry"] = security.SubIndustry.String
	}
	if security.Consensus.Valid {
		updates = append(updates, "consensus = :consensus")
		args["consensus"] = security.Consensus.String
	}
	if security.STM.Valid {
		updates = append(updates, "stm = :stm")
		args["stm"] = security.STM.String
	}

	// Integer Fields (only update if non-zero)
	if security.Price != 0 {
		updates = append(updates, "price = :price")
		args["price"] = security.Price
	}
	if security.PC != 0 {
		updates = append(updates, "pc = :pc")
		args["pc"] = security.PC
	}
	if security.PCP != 0 {
		updates = append(updates, "pcp = :pcp")
		args["pcp"] = security.PCP
	}
	if security.YearLow != 0 {
		updates = append(updates, "yrl = :yrl")
		args["yrl"] = security.YearLow
	}
	if security.YearHigh != 0 {
		updates = append(updates, "yrh = :yrh")
		args["yrh"] = security.YearHigh
	}
	if security.DayLow != 0 {
		updates = append(updates, "drl = :drl")
		args["drl"] = security.DayLow
	}
	if security.DayHigh != 0 {
		updates = append(updates, "drh = :drh")
		args["drh"] = security.DayHigh
	}
	if security.PClose != 0 {
		updates = append(updates, "pclose = :pclose")
		args["pclose"] = security.PClose
	}
	if security.COpen != 0 {
		updates = append(updates, "copen = :copen")
		args["copen"] = security.COpen
	}
	if security.Bid != 0 {
		updates = append(updates, "bid = :bid")
		args["bid"] = security.Bid
	}
	if security.Ask != 0 {
		updates = append(updates, "ask = :ask")
		args["ask"] = security.Ask
	}

	// Nullable Int Fields
	if security.Score.Valid {
		updates = append(updates, "score = :score")
		args["score"] = security.Score.Int64
	}
	if security.Coverage.Valid {
		updates = append(updates, "coverage = :coverage")
		args["coverage"] = security.Coverage.Int64
	}
	if security.MarketCap.Valid {
		updates = append(updates, "cap = :cap")
		args["cap"] = security.MarketCap.Int64
	}
	if security.Volume.Valid {
		updates = append(updates, "volume = :volume")
		args["volume"] = security.Volume.Int64
	}
	if security.AvgVolume.Valid {
		updates = append(updates, "avgvolume = :avgvolume")
		args["avgvolume"] = security.AvgVolume.Int64
	}
	if security.Outstanding.Valid {
		updates = append(updates, "outstanding = :outstanding")
		args["outstanding"] = security.Outstanding.Int64
	}
	if security.Beta.Valid {
		updates = append(updates, "beta = :beta")
		args["beta"] = security.Beta.Int64
	}
	if security.BidSize.Valid {
		updates = append(updates, "bidsz = :bidsz")
		args["bidsz"] = security.BidSize.Int64
	}
	if security.AskSize.Valid {
		updates = append(updates, "asksz = :asksz")
		args["asksz"] = security.AskSize.Int64
	}
	if security.EPS.Valid {
		updates = append(updates, "eps = :eps")
		args["eps"] = security.EPS.Int64
	}
	if security.PE.Valid {
		updates = append(updates, "pe = :pe")
		args["pe"] = security.PE.Int64
	}
	if security.Target.Valid {
		updates = append(updates, "target = :target")
		args["target"] = security.Target.Int64
	}

	// Ensure at least one field is being updated
	if len(updates) == 0 {
		return fmt.Errorf("no valid fields to update for ticker: %s", security.Ticker)
	}

	// Final query construction
	query += strings.Join(updates, ", ") + " WHERE ticker = :ticker"
	args["ticker"] = security.Ticker

	// Execute update query
	_, err := tx.NamedExec(query, args)
	if err != nil {
		return fmt.Errorf("failed to update security (%s): %w", security.Ticker, err)
	}

	return nil
}

func SecurityExists(db *sqlx.DB, ticker string, exchange string) bool {
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM securities WHERE ticker = $1 AND exchange = $2)", ticker, exchange)
	if err != nil {
		return false
	}
	return exists
}

func GetSecurityView(db *sqlx.DB, tp, input string) (*SelectedSecurityView, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected: ticker:exchange")
	}
	ticker, exchange := parts[0], parts[1]
	tp = strings.ToUpper(tp)

	query := `
		SELECT
			s.ticker, s.exchange, s.fullname,
			s.price, s.typology, s.currency, s.target,

			COALESCE(d.yield, 0),
    		d.ap , d.pr,
     		d.frequency,

			COALESCE(e.family, ''), e.er

		FROM securities s
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		LEFT JOIN etfs e ON s.ticker = e.ticker AND s.exchange = e.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange AND s.typology = :tp
	`

	// Execute the query using NamedQuery
	rows, err := db.NamedQuery(query, map[string]any{
		"ticker":   ticker,
		"exchange": exchange,
		"tp":       tp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve selected security '%s': %w", input, err)
	}
	defer rows.Close()

	// Check if any rows were returned
	if !rows.Next() {
		return nil, fmt.Errorf("selected security '%s' not found", input)
	}

	// Parse result into Security struct
	var selectedSecurity SelectedSecurityView
	if err := selectedSecurity.Scan(rows); err != nil {
		return nil, fmt.Errorf("failed to scan selected security '%s': %w", input, err)
	}

	return &selectedSecurity, nil
}

func GetSecurityVars(db *sqlx.DB, input string) (*SecurityVars, error) {
	// Parse the input into ticker and exchange
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format, expected: ticker:exchange")
	}
	ticker, exchange := parts[0], parts[1]

	query := `
		SELECT
			s.price, s.currency,

			COALESCE(d.yield, 0),
     		d.frequency,
			e.er, d.pd

		FROM securities s
		LEFT JOIN dividends d ON s.ticker = d.ticker AND s.exchange = d.exchange
		LEFT JOIN etfs e ON s.ticker = e.ticker AND s.exchange = e.exchange
		WHERE s.ticker = :ticker AND s.exchange = :exchange
	`

	// Execute the query using NamedQuery
	rows, err := db.NamedQuery(query, map[string]any{
		"ticker":   ticker,
		"exchange": exchange,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve selected security '%s': %w", input, err)
	}
	defer rows.Close()

	// Check if any rows were returned
	if !rows.Next() {
		return nil, fmt.Errorf("selected security '%s' not found", input)
	}

	// Parse result into Security struct
	var vars SecurityVars
	if err := vars.Scan(rows); err != nil {
		return nil, fmt.Errorf("failed to scan selected security '%s': %w", input, err)
	}

	return &vars, nil
}
