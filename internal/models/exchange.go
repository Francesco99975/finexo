package models

import (
	"fmt"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/jmoiron/sqlx"
)

type Exchange struct {
	Title     string         `db:"title" json:"title"`
	Fullname  string         `db:"fullname" json:"fullname"`
	Prefix    NullableString `db:"prefix" json:"prefix,omitempty"`
	Suffix    NullableString `db:"suffix" json:"suffix,omitempty"`
	CC        string         `db:"cc" json:"countryCode"`
	OpenTime  NullableTime   `db:"opentime" json:"openTime,omitempty"`
	CloseTime NullableTime   `db:"closetime" json:"closeTime,omitempty"`
}

func InitExchanges(db *sqlx.DB) ([]Exchange, error) {

	exchanges := []Exchange{
		// NYSE
		{
			Title:    "NYSE",
			Fullname: "New York Stock Exchange",
			CC:       "US",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 21, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// NASDAQ
		{
			Title:    "NASDAQ",
			Fullname: "National Association of Securities Dealers Automated Quotations",
			CC:       "US",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 21, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// TSX
		{
			Title:    "TSX",
			Fullname: "Toronto Stock Exchange",
			Prefix: NullableString{
				String: "TSE",
				Valid:  true,
			},
			Suffix: NullableString{
				String: "TO",
				Valid:  true,
			},
			CC: "CA",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 21, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// TSXV
		{
			Title:    "TSXV",
			Fullname: "TSX Venture Exchange",
			Prefix: NullableString{
				String: "CVE",
				Valid:  true,
			},
			Suffix: NullableString{
				String: "V",
				Valid:  true,
			},
			CC: "CA",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 21, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// CBOE CA
		{
			Title:    "CBOE",
			Fullname: "CBOE Canada",
			Prefix: NullableString{
				String: "NEOA",
				Valid:  true,
			},
			Suffix: NullableString{
				String: "NE",
				Valid:  true,
			},
			CC: "CA",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 21, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// CBOE US
		{
			Title:    "CBOEUS",
			Fullname: "Chicago Board Options Exchange",
			CC:       "US",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 21, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// London Stock Exchange (LSE)
		{
			Title:    "LSE",
			Fullname: "London Stock Exchange",
			Prefix: NullableString{
				String: "LON",
				Valid:  true,
			},
			Suffix: NullableString{
				String: "L",
				Valid:  true,
			},
			CC: "GB",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 30, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// Milan Stock Exchange (MIL)
		{
			Title:    "MIL",
			Fullname: "Milan Stock Exchange",
			CC:       "IT",
			Suffix: NullableString{
				String: "MI",
				Valid:  true,
			},
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 30, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// Tokyo Stock Exchange (Tokyo)
		{
			Title:    "JPY",
			Fullname: "Tokyo Stock Exchange",
			Suffix: NullableString{
				String: "T",
				Valid:  true,
			},
			CC: "JP",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 6, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// Frankfurt Stock Exchange (FWB)
		{
			Title:    "FWB",
			Fullname: "Frankfurt Stock Exchange",
			CC:       "DE",
			Prefix: NullableString{
				String: "FRA",
				Valid:  true,
			},
			Suffix: NullableString{
				String: "F",
				Valid:  true,
			},
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 30, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// SIX Swiss Exchange (SIX) - New
		{
			Title:    "SIX",
			Fullname: "SIX Swiss Exchange",
			Suffix: NullableString{
				String: "SW",
				Valid:  true,
			},
			CC: "CH",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 30, 0, 0, time.UTC),
				Valid: true,
			},
		},
		// Australian Securities Exchange (ASX) - New
		{
			Title:    "ASX",
			Fullname: "Australian Securities Exchange",
			Suffix: NullableString{
				String: "AX",
				Valid:  true,
			},
			CC: "AU",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 6, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
	}

	for _, exchange := range exchanges {
		err := CreateExchange(db, &exchange)
		if err != nil {
			return nil, fmt.Errorf("failed to create exchange: %w", err)
		}
	}

	return exchanges, nil
}

func CreateExchange(db *sqlx.DB, exchange *Exchange) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer database.HandleTransaction(tx, &err)

	// Insert into exchanges table
	query := `
		INSERT INTO exchanges (title, fullname, prefix, suffix, cc, opentime, closetime) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (title) DO NOTHING
	`

	_, err = tx.Exec(query, exchange.Title, exchange.Title, exchange.Prefix, exchange.Suffix, exchange.CC, exchange.OpenTime, exchange.CloseTime)

	if err != nil {
		return fmt.Errorf("failed to insert exchange: %w", err)
	}

	return nil
}

func GetExchangeBySuffixorPrefix(db *sqlx.DB, suffix, prefix string) (*Exchange, error) {
	query := `
		SELECT title, fullname, prefix, suffix, cc, opentime, closetime
		FROM exchanges
		WHERE (suffix = $1 OR prefix = $2)
	`
	var exchange Exchange
	err := db.Get(&exchange, query, suffix, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange: %w", err)
	}
	return &exchange, nil
}

func GetExchangeByTitle(db *sqlx.DB, title string) (*Exchange, error) {
	query := `
		SELECT title, fullname, prefix, suffix, cc, opentime, closetime
		FROM exchanges
		WHERE title = $1
	`
	var exchange Exchange
	err := db.Get(&exchange, query, title)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange: %w", err)
	}
	return &exchange, nil
}

func GetAllExchanges(db *sqlx.DB) ([]Exchange, error) {
	query := `
		SELECT title, fullname, prefix, suffix, cc, opentime, closetime
		FROM exchanges
	`
	var exchanges []Exchange
	err := db.Select(&exchanges, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchanges: %w", err)
	}
	return exchanges, nil
}
