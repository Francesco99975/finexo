package api

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func GetStocks() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.SecParamsPointers
		err := c.Bind(&payload)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid stocks request payload", Error: err.Error()})
		}

		params, err := payload.Validate()
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Validation Error", Error: err.Error()})
		}

		stocks, err := models.GetStocks(database.DB, params.Exchange, params.Country, params.MinPrice, params.MaxPrice, params.Order, params.Asc, params.Limit, params.Dividend)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve stocks", Error: err.Error()})
		}

		if len(stocks) == 0 {
			return c.JSON(http.StatusNotFound, models.JSONErrorResponse{Code: http.StatusNotFound, Message: "No matching stocks found", Error: "No matching stocks found"})
		}

		return c.JSON(http.StatusOK, stocks)
	}
}

func GetStock() echo.HandlerFunc {
	return func(c echo.Context) error {
		stock, err := models.GetStock(database.DB, c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve stock", Error: err.Error()})
		}

		return c.JSON(http.StatusOK, stock)
	}
}
