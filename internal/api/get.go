package api

import (
	"fmt"
	"net/http"

	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/echo/v4"
)

func PlaceholderGet() echo.HandlerFunc {
	return func(c echo.Context) error {
		categories, err := models.PlaceholderGet()
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: fmt.Sprintf("Error fetching <- %v", err), Errors: []string{err.Error()}})
		}

		return c.JSON(http.StatusOK, categories)
	}
}
