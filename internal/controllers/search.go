package controllers

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views/components"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func SearchHtmlSecurities() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Grab seed from params
		query := c.QueryParam("q")
		if query == "" {
			log.Warn("No query provided")

			html := helpers.MustRenderHTML(components.WarnMsg("No query provided"))
			return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
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
			log.Errorf("Could not query database: %w", err)

			html := helpers.MustRenderHTML(components.ErrorMsg("Could not query database"))
			return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
		}
		defer rows.Close()

		var seachResults []models.SecuritySearchView
		for rows.Next() {
			var sec models.SecuritySearchView
			if err := sec.Scan(rows); err == nil {
				seachResults = append(seachResults, sec)
			}
		}

		html := helpers.MustRenderHTML(components.SearchSecurityItems(seachResults))

		return c.Blob(200, "text/html; charset=utf-8", html)

	}
}
