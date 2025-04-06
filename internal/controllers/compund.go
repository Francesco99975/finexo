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
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid form data: %w", err))
		}

		var results helpers.CalculationResults
		if input.SID != "default" {
			vars, err := models.GetSecurityVars(database.DB, input.SID)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("data not found in databse %w", err))
			}

			results, err = helpers.CalculateInvestment(input.SID, vars.Price, vars.Yield, vars.ExpenseRatio, input.Principal, input.Contribution, input.ContribFrequency, vars.Frequency, input.PriceMod, input.YieldMod, input.Years, vars.PayoutMonth, vars.Currency)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Could not calculate investment compound: %w", err))
			}
		} else {
			var err error
			results, err = helpers.CalculateHISAInvestment(input.Principal, input.Contribution, input.ContribFrequency, input.CompundingFrequency, input.Rate, input.Years, input.Currency)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Could not calculate hisa compound: %w", err))
			}
		}

		encodedResults, err := results.Encoded()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Could not encode results: %w", err))
		}

		csrfToken := c.Get("csrf").(string)

		html, err := helpers.RenderHTML(components.Calculations(results, encodedResults, csrfToken))

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse page home")
		}

		return c.Blob(200, "text/html; charset=utf-8", html)

	}
}
