package api

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func GetStocks() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.SecParams
		err := c.Bind(&payload)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid stocks request payload"})
		}

		err = payload.Validate()
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}

		stocks, err := models.GetStocks(database.DB, payload.Exchange, payload.Country, payload.MinPrice, payload.MaxPrice, payload.Order, payload.Asc, payload.Limit, payload.Dividend != nil && *payload.Dividend)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve stocks"})
		}

		return c.JSON(http.StatusOK, stocks)
	}
}

func GetStock() echo.HandlerFunc {
	return func(c echo.Context) error {
		stock, err := models.GetStock(database.DB, c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve stock"})
		}

		return c.JSON(http.StatusOK, stock)
	}
}
