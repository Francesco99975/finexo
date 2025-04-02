package controllers

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views/components"
	"github.com/labstack/echo/v4"
)

func SearchHtmlSecurities() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Grab seed from params
		query := c.QueryParam("q")
		if query == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid query")
		}

		// Perform the search with trigram similarity ordering
		rows, err := database.DB.Queryx(`
		SELECT ticker, exchange, fullname, price, typology, currency
		FROM securities
		WHERE fullname ILIKE '%' || $1 || '%'
		   OR ticker ILIKE '%' || $1 || '%'
		ORDER BY similarity(fullname, $1) DESC
		LIMIT 10`, query)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to seach db")
		}
		defer rows.Close()

		var seachResults []models.SecuritySearchView
		for rows.Next() {
			var sec models.SecuritySearchView
			if err := sec.Scan(rows); err == nil {
				seachResults = append(seachResults, sec)
			}
		}

		html, err := helpers.RenderHTML(components.SearchSecurityItems(seachResults))

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse page home")
		}

		return c.Blob(200, "text/html; charset=utf-8", html)

	}
}
