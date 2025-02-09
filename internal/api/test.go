package api

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Test() echo.HandlerFunc {
	return func(c echo.Context) error {

		log.Info("Test endpoint called")

		err := tools.Scrape("QQQ", nil)
		if err != nil {
			log.Errorf("Failed to scrape: %v", err)
		}

		return c.JSON(http.StatusOK, "OK")
	}
}
