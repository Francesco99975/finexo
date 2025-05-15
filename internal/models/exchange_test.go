package models

import (
	"reflect"
	"testing"

	"github.com/78bits/go-sqlmock-sqlx"
	_ "github.com/lib/pq"
)

func TestGetAllTickersFromExchanges(t *testing.T) {
	// 1) Create mock DB
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("unexpected error opening stub database: %s", err)
	}
	defer db.Close()

	// 2) Prepare the rows we want returned:
	//    note the column name must match the CONCAT(...) label
	rows := sqlmock.NewRows([]string{"CONCAT(ticker, ':', exchange)"})
	rows.
		AddRow("KAG:ASX").
		AddRow("6969:JPY")

	// 3) Expect exactly the SELECT CONCAT(...) query—with placeholders—and the two args
	mock.
		ExpectQuery(`SELECT CONCAT\(ticker, ':', exchange\) FROM securities WHERE exchange IN \(\?, \?\)`).
		WithArgs("ASX", "JPY").
		WillReturnRows(rows)

	// 4) Call the function under test
	tickers, err := GetAllTickersFromExchanges(db, []string{"ASX", "JPY"})
	if err != nil {
		t.Fatalf("unexpected error from GetAllTickersFromExchanges: %v", err)
	}

	// 5) Verify results
	want := []string{"KAG:ASX", "6969:JPY"}
	if !reflect.DeepEqual(tickers, want) {
		t.Errorf("tickers = %v; want %v", tickers, want)
	}

	// 6) Ensure no other mocks were left hanging
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sqlmock expectations: %s", err)
	}
}
