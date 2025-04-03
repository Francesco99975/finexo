package controllers

import (
	"fmt"
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views/components"
	"github.com/labstack/echo/v4"
)

func CalculateCompound() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input models.CalcInput
		if err := c.Bind(&input); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		}

		var results helpers.CalculationResults
		if input.SID != "" {
			vars, err := models.GetSecurityVars(database.DB, input.SID)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
			}

			results, err = helpers.CalculateInvestment(vars.Price, vars.Yield, vars.ExpenseRatio, input.Principal, input.Contribution, input.ContribFrequency, vars.Frequency, input.PriceMod, input.YieldMod, input.Years, vars.PayoutMonth, vars.Currency)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Could not calculate: %w", err))
			}
		}

		html, err := helpers.RenderHTML(components.Calculations(results))

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse page home")
		}

		return c.Blob(200, "text/html; charset=utf-8", html)

	}
}
