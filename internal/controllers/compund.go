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

func CalculateCompound() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input models.CalcInput
		if err := c.Bind(&input); err != nil {
			log.Errorf("failed to bind form data: %w", err)

			html := helpers.MustRenderHTML(components.ErrorMsg("Invalid form data"))

			return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
		}

		var results helpers.CalculationResults
		if input.SID != "default" {
			vars, err := models.GetSecurityVars(database.DB, input.SID)
			if err != nil {
				log.Errorf("failed to get security vars: %w", err)

				html := helpers.MustRenderHTML(components.ErrorMsg("Could not get identify security to do calculations"))

				return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
			}

			results, err = helpers.CalculateInvestment(input.SID, vars.Price, vars.Yield, vars.ExpenseRatio, input.Principal, input.Contribution, input.ContribFrequency, vars.Frequency, input.PriceMod, input.YieldMod, input.Years, vars.PayoutMonth, vars.Currency)
			if err != nil {
				log.Errorf("failed to calculate investment compound: %w", err)

				html := helpers.MustRenderHTML(components.ErrorMsg("Failed to calculate investment compound"))
				return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
			}
			helpers.RecordBusinessEvent("calculate_investment")
		} else {
			var err error
			results, err = helpers.CalculateHISAInvestment(input.Principal, input.Contribution, input.ContribFrequency, input.CompundingFrequency, input.Rate, input.Years, input.Currency)
			if err != nil {
				log.Errorf("failed to calculate hisa compound: %w", err)

				html := helpers.MustRenderHTML(components.ErrorMsg("Failed to calculate hisa compound"))
				return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
			}
			helpers.RecordBusinessEvent("calculate_hisa")
		}

		encodedResults, err := results.Encoded()
		if err != nil {
			log.Errorf("failed to encode results: %w", err)
			html := helpers.MustRenderHTML(components.ErrorMsg("Failed to encode results"))
			return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
		}

		csrfToken := c.Get("csrf").(string)

		html := helpers.MustRenderHTML(components.Calculations(results, encodedResults, csrfToken))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)

	}
}
