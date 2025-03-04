package api

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func GetETFs() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.SecParamsPointers
		err := c.Bind(&payload)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid etfs request payload"})
		}

		params, err := payload.Validate()
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}

		etfs, err := models.GetETFs(database.DB, params.Exchange, params.Country, params.MinPrice, params.MaxPrice, params.Order, params.Asc, params.Limit, params.Dividend)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve etfs"})
		}

		return c.JSON(http.StatusOK, etfs)
	}
}

func GetETF() echo.HandlerFunc {
	return func(c echo.Context) error {
		etf, err := models.GetETF(database.DB, c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve etf"})
		}

		return c.JSON(http.StatusOK, etf)
	}
}
