package controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/echo/v4"
)

func DownloadPDF() echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedResults := c.Param("results")
		if encodedResults == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "No results found")
		}

		results, err := helpers.DecodeResults(encodedResults)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Could not decode results: %w", err))
		}

		filename, err := tools.GeneratePDF(results)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Could not generate PDF: %w", err))
		}

		defer os.Remove(filename)

		return c.Attachment(filename, filename)

	}
}
