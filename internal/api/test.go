package api

import (
	"database/sql"
	"net/http"

	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Test() echo.HandlerFunc {
	return func(c echo.Context) error {

		log.Info("Test endpoint called")

		err := tools.Scrape("QQQ", models.Exchange{Title: "NASDAQ", Suffix: sql.NullString{Valid: false}, Prefix: sql.NullString{Valid: false}, CC: "US"}, models.Country{Code: "US", Label: "United States", Currency: "USD", Continent: "America", ISO: "US"})
		if err != nil {
			log.Errorf("Failed to scrape: %v", err)
		}

		return c.JSON(http.StatusOK, "")
	}
}
