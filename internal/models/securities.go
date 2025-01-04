package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
)

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
			:ticker, :cc, :suffix, :exchange, 'etf', :fullname, :price, :pc, :ppc,
			:yrange, :drange, :marketcap, :volume, :avgvlm, :beta, :pclose, :copen, :bid, :bidsz,
			:ask, :asksz, :currency, :created, :updated
		)
	`
	_, err = tx.NamedExec(securitiesQuery, &etf.Security)
	if err != nil {
		return fmt.Errorf("failed to insert security: %w", err)
	}

	// Insert into etfs table
	etfsQuery := `
		INSERT INTO etfs (ticker, cc, holdings, aum, er)
		VALUES (:ticker, :cc, :holdings, :aum, :expenseRatio)
	`
	_, err = tx.NamedExec(etfsQuery, etf)
	if err != nil {
		return fmt.Errorf("failed to insert etf: %w", err)
	}

	// Insert related securities
	if len(etf.RelatedSecurities) > 0 {
		relatedQuery := `
			INSERT INTO etf_related_securities (etf_ticker, etf_cc, related_ticker, related_cc)
			VALUES (:etf_ticker, :etf_cc, :related_ticker, :related_cc)
		`
		for _, related := range etf.RelatedSecurities {
			_, err = tx.NamedExec(relatedQuery, map[string]interface{}{
				"etf_ticker":     etf.Ticker,
				"etf_cc":         etf.Country,
				"related_ticker": related.Ticker,
				"related_cc":     related.Country,
			})
			if err != nil {
				return fmt.Errorf("failed to insert related security: %w", err)
			}
		}
	}

	// Insert into dividends table if provided
	if etf.Security.Dividend != nil {
		dividendsQuery := `
			INSERT INTO dividends (
				ticker, cc, rate, trate, yield, tyield, ap, tap, pr, lgr, yog, lad, frequency, edd, pd
			) VALUES (
				:ticker, :cc, :rate, :rateType, :yield, :yieldType, :annualPayout, :apType,
				:payoutRatio, :growthRate, :yearsGrowth, :lastAnnounced, :frequency, :exDivDate, :payoutDate
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
			dividend.Country = dividendData["cc"].(string)
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
	Country       string         `db:"cc" json:"country"`
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
	RelatedSecurities []Security        `json:"relatedSecurities"` // Related securities for the ETF
}

// REIT represents a row from the reits table.
type REIT struct {
	Security `json:"security"` // Embedded security properties
	FFO      sql.NullInt64     `db:"ffo" json:"ffo,omitempty"`
	FFOType  sql.NullString    `db:"tffo" json:"ffoType,omitempty"`
	PFFO     sql.NullInt64     `db:"pffo" json:"pffo,omitempty"`
	PFFOType sql.NullString    `db:"tpffo" json:"pffoType,omitempty"`
}
