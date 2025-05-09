package api

import (
	"net/http"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func GetETFs() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.SecParamsPointers
		err := c.Bind(&payload)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid etfs request payload", Error: err.Error()})
		}

		params, err := payload.Validate()
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Validation Error", Error: err.Error()})
		}

		start := time.Now()
		etfs, err := models.GetETFs(database.DB, params.Exchange, params.Country, params.MinPrice, params.MaxPrice, params.Order, params.Asc, params.Limit, params.Dividend)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve etfs", Error: err.Error()})
		}
		helpers.RecordDBQueryLatency("get_etfs", start)
		helpers.RecordBusinessEvent("get_etfs")

		if len(etfs) == 0 {
			return c.JSON(http.StatusNotFound, models.JSONErrorResponse{Code: http.StatusNotFound, Message: "No matching etfs found", Error: "No matching etfs found"})
		}

		return c.JSON(http.StatusOK, etfs)
	}
}

func GetETF() echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		etf, err := models.GetETF(database.DB, c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve etf", Error: err.Error()})
		}
		helpers.RecordDBQueryLatency("get_etf", start)
		helpers.RecordBusinessEvent("get_etf")

		return c.JSON(http.StatusOK, etf)
	}
}
