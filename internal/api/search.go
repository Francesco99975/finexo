package api

import (
	"net/http"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func SearchSecurities() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Grab seed from params
		query := c.QueryParam("q")
		if query == "" {
			return c.JSON(http.StatusBadRequest, "Missing seach query parameter")
		}

		start := time.Now()

		// Perform the search with trigram similarity ordering
		rows, err := database.DB.Queryx(`
		SELECT ticker, exchange, fullname, price, typology
		FROM securities
		WHERE fullname ILIKE '%' || $1 || '%'
		   OR ticker ILIKE '%' || $1 || '%'
		ORDER BY similarity(fullname, $1) DESC
		LIMIT 10`, query)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to seach db", Error: err.Error()})
		}
		defer rows.Close()

		var seachResults []models.SecuritySearchView
		for rows.Next() {
			var sec models.SecuritySearchView
			if err := sec.Scan(rows); err == nil {
				seachResults = append(seachResults, sec)
			}
		}

		helpers.RecordDBQueryLatency("search_securities", start)
		helpers.RecordBusinessEvent("search_securities")

		return c.JSON(http.StatusOK, seachResults)

	}
}
