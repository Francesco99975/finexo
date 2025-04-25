package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Dividend represents a row from the dividends table.
type Dividend struct {
	Ticker        string         `db:"ticker" json:"ticker,omitempty"`
	Exchange      string         `db:"exchange" json:"exchange,omitempty"`
	Yield         int            `db:"yield" json:"yield"`
	AnnualPayout  NullableInt    `db:"ap" json:"annualPayout,omitempty"`
	Timing        NullableString `db:"tm" json:"timing,omitempty"` // Enum: fwd, ttm
	PayoutRatio   NullableInt    `db:"pr" json:"payoutRatio,omitempty"`
	GrowthRate    NullableInt    `db:"lgr" json:"growthRate,omitempty"`
	YearsGrowth   NullableInt    `db:"yog" json:"yearsGrowth,omitempty"`
	LastAnnounced NullableInt    `db:"lad" json:"lastAnnounced,omitempty"`
	Frequency     NullableString `db:"frequency" json:"frequency,omitempty"` // Enum: Frequency
	ExDivDate     NullableTime   `db:"edd" json:"exDivDate,omitempty"`
	PayoutDate    NullableTime   `db:"pd" json:"payoutDate,omitempty"`
}

func (d *Dividend) PrettyPrintString() string {
	var sb strings.Builder
	sb.WriteString("Yield: " + strconv.Itoa(d.Yield) + " -- ")
	if d.AnnualPayout.Valid {
		sb.WriteString("Annual Payout: " + strconv.Itoa(int(d.AnnualPayout.Int64)) + " -- ")
	}
	if d.Timing.Valid {
		sb.WriteString("Timing: " + d.Timing.String + " -- ")
	}
	if d.PayoutRatio.Valid {
		sb.WriteString("Payout Ratio: " + strconv.Itoa(int(d.PayoutRatio.Int64)) + " -- ")
	}
	if d.GrowthRate.Valid {
		sb.WriteString("Growth Rate: " + strconv.Itoa(int(d.GrowthRate.Int64)) + " -- ")
	}
	if d.YearsGrowth.Valid {
		sb.WriteString("Years Growth: " + strconv.Itoa(int(d.YearsGrowth.Int64)) + " -- ")
	}
	if d.LastAnnounced.Valid {
		sb.WriteString("Last Announced: " + strconv.Itoa(int(d.LastAnnounced.Int64)) + " -- ")
	}
	if d.Frequency.Valid {
		sb.WriteString("Frequency: " + d.Frequency.String + " -- ")
	}
	if d.ExDivDate.Valid {
		sb.WriteString("Ex-Div Date: " + d.ExDivDate.Time.Format("2006-01-02") + " -- ")
	}
	if d.PayoutDate.Valid {
		sb.WriteString("Payout Date: " + d.PayoutDate.Time.Format("2006-01-02") + " -- ")
	}
	return sb.String()
}

func InsertDividend(tx *sqlx.Tx, dividend *Dividend) error {
	if dividend == nil {
		return nil // No dividend, skip insertion
	}

	query := `
		INSERT INTO dividends (
			ticker, exchange, yield, ap, tm, pr, lgr, yog, lad, frequency, edd, pd
		) VALUES (
			:ticker, :exchange, :yield, :ap, :tm, :pr,
			:lgr, :yog, :lad, :frequency, :edd, :pd
		)
	`

	_, err := tx.NamedExec(query, dividend)
	if err != nil {
		return fmt.Errorf("failed to insert dividend for %s (%s): %w", dividend.Ticker, dividend.Exchange, err)
	}

	return nil
}

func UpdateDividend(tx *sqlx.Tx, dividend *Dividend) error {
	if dividend == nil {
		return nil // No dividend to update
	}

	if dividend.Ticker == "" || dividend.Exchange == "" {
		return fmt.Errorf("ticker and exchange are required for updating a dividend")
	}

	query := "UPDATE dividends SET "
	args := make(map[string]any)
	updates := []string{}

	// Integer Fields
	if dividend.Yield != 0 {
		updates = append(updates, "yield = :yield")
		args["yield"] = dividend.Yield
	}

	// Nullable Integer Fields
	if dividend.AnnualPayout.Valid {
		updates = append(updates, "ap = :ap")
		args["ap"] = dividend.AnnualPayout.Int64
	}
	if dividend.PayoutRatio.Valid {
		updates = append(updates, "pr = :pr")
		args["pr"] = dividend.PayoutRatio.Int64
	}
	if dividend.GrowthRate.Valid {
		updates = append(updates, "lgr = :lgr")
		args["lgr"] = dividend.GrowthRate.Int64
	}
	if dividend.YearsGrowth.Valid {
		updates = append(updates, "yog = :yog")
		args["yog"] = dividend.YearsGrowth.Int64
	}
	if dividend.LastAnnounced.Valid {
		updates = append(updates, "lad = :lad")
		args["lad"] = dividend.LastAnnounced.Int64
	}

	// Nullable String Fields
	if dividend.Timing.Valid {
		updates = append(updates, "tm = :tm")
		args["tm"] = dividend.Timing.String
	}
	if dividend.Frequency.Valid {
		updates = append(updates, "frequency = :frequency")
		args["frequency"] = dividend.Frequency.String
	}

	// Nullable Time Fields
	if dividend.ExDivDate.Valid {
		updates = append(updates, "edd = :edd")
		args["edd"] = dividend.ExDivDate.Time
	}
	if dividend.PayoutDate.Valid {
		updates = append(updates, "pd = :pd")
		args["pd"] = dividend.PayoutDate.Time
	}

	// Ensure at least one field is being updated
	if len(updates) == 0 {
		return fmt.Errorf("no valid fields to update for dividend with ticker: %s", dividend.Ticker)
	}

	// Final query construction
	query += strings.Join(updates, ", ") + " WHERE ticker = :ticker AND exchange = :exchange"
	args["ticker"] = dividend.Ticker
	args["exchange"] = dividend.Exchange

	// Execute update query
	_, err := tx.NamedExec(query, args)
	if err != nil {
		return fmt.Errorf("failed to update dividend (%s): %w", dividend.Ticker, err)
	}

	return nil
}
