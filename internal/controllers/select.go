package controllers

import (
	"fmt"
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views/components"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Select() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Grab seed from params
		tp := c.Param("tp")
		id := c.Param("id")
		if id == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
		}

		// Check if tp is valid
		if tp != "stock" && tp != "etf" && tp != "reit" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid typology")
		}

		// Get security from db
		selectedSecurity, err := models.GetSecurityView(database.DB, tp, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Could not get security from db: %s", err))
		}

		log.Infof("Selected security: %+v", selectedSecurity)

		html, err := helpers.RenderHTML(components.SelectedSecurity(*selectedSecurity))

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse page home")
		}

		return c.Blob(200, "text/html; charset=utf-8", html)

	}
}
