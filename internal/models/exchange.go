package models

import (
	"fmt"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/jmoiron/sqlx"
)

type Exchange struct {
	Title     string         `db:"title" json:"title"`
	Prefix    NullableString `db:"prefix" json:"prefix,omitempty"`
	Suffix    NullableString `db:"suffix" json:"suffix,omitempty"`
	CC        string         `db:"cc" json:"countryCode"`
	OpenTime  NullableTime   `db:"opentime" json:"openTime,omitempty"`
	CloseTime NullableTime   `db:"closetime" json:"closeTime,omitempty"`
}

func InitExchanges(db *sqlx.DB) error {

	exchanges := []Exchange{
		{
			Title: "NYSE",
			CC:    "US",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 9, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		{
			Title: "NASDAQ",
			CC:    "US",
			OpenTime: NullableTime{
				Time:  time.Date(0, 0, 0, 9, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		{
			Title: "TSX",
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
				Time:  time.Date(0, 0, 0, 9, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		{
			Title: "TSXV",
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
				Time:  time.Date(0, 0, 0, 9, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
		{
			Title: "CBOE",
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
				Time:  time.Date(0, 0, 0, 9, 30, 0, 0, time.UTC),
				Valid: true,
			},
			CloseTime: NullableTime{
				Time:  time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
	}

	for _, exchange := range exchanges {
		err := CreateExchange(db, &exchange)
		if err != nil {
			return fmt.Errorf("failed to create exchange: %w", err)
		}
	}

	return nil
}

func CreateExchange(db *sqlx.DB, exchange *Exchange) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer database.HandleTransaction(tx, &err)

	// Insert into exchanges table
	query := `
		INSERT INTO exchanges (title, prefix, suffix, cc, opentime, closetime) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (title) DO NOTHING
	`

	_, err = tx.Exec(query, exchange.Title, exchange.Prefix, exchange.Suffix, exchange.CC, exchange.OpenTime, exchange.CloseTime)

	if err != nil {
		return fmt.Errorf("failed to insert exchange: %w", err)
	}

	return nil
}

func GetExchangeBySuffixorPrefix(db *sqlx.DB, suffix, prefix string) (*Exchange, error) {
	query := `
		SELECT title, prefix, suffix, cc, opentime, closetime
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
		SELECT title, prefix, suffix, cc, opentime, closetime
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
