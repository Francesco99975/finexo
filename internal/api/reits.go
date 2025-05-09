package api

import (
	"net/http"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func GetREITs() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.SecParamsPointers
		err := c.Bind(&payload)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid reits request payload", Error: err.Error()})
		}

		params, err := payload.Validate()
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}

		start := time.Now()

		reits, err := models.GetREITs(database.DB, params.Exchange, params.Country, params.MinPrice, params.MaxPrice, params.Order, params.Asc, params.Limit, params.Dividend)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve reits", Error: err.Error()})
		}

		helpers.RecordDBQueryLatency("get_reits", start)
		helpers.RecordBusinessEvent("get_reits")

		if len(reits) == 0 {
			return c.JSON(http.StatusNotFound, models.JSONErrorResponse{Code: http.StatusNotFound, Message: "No matching reits found", Error: "No matching reits found"})
		}

		return c.JSON(http.StatusOK, reits)
	}
}

func GetREIT() echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		reit, err := models.GetREIT(database.DB, c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve reit"})
		}
		helpers.RecordDBQueryLatency("get_reit", start)
		helpers.RecordBusinessEvent("get_reit")

		return c.JSON(http.StatusOK, reit)
	}
}
