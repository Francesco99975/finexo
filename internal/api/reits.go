package api

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func GetREITs() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.SecParams
		err := c.Bind(&payload)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid reits request payload"})
		}

		err = payload.Validate()
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}

		reits, err := models.GetREITs(database.DB, payload.Exchange, payload.Country, payload.MinPrice, payload.MaxPrice, payload.Order, payload.Asc, payload.Limit, payload.Dividend != nil && *payload.Dividend)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve reits"})
		}

		return c.JSON(http.StatusOK, reits)
	}
}

func GetREIT() echo.HandlerFunc {
	return func(c echo.Context) error {
		reit, err := models.GetETF(database.DB, c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve reit"})
		}

		return c.JSON(http.StatusOK, reit)
	}
}
