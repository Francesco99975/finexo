package tools

import (
	"database/sql"
	"testing"

	"github.com/Francesco99975/finexo/internal/models"
)

// TestScrape tests the scrape function
func TestScrape(t *testing.T) {

	// Call scrape() with the mock server URL
	err := Scrape("AAPL", models.Exchange{Title: "NASDAQ", Suffix: sql.NullString{Valid: false}, Prefix: sql.NullString{Valid: false}, CC: "US"}, models.Country{Code: "US", Label: "United States", Currency: "USD", Continent: "America", ISO: "US"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
