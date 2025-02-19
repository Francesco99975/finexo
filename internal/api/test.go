package api

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Test() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Grab seed from params
		seed := c.Param("seed")
		if seed == "" {
			return c.JSON(http.StatusBadRequest, "Missing seed parameter")
		}
		log.Info("Test endpoint called")

		err := tools.Scrape(seed, nil)
		if err != nil {
			log.Errorf("Failed to scrape: %v", err)
		}

		return c.JSON(http.StatusOK, "OK")
	}
}

func TestSeeds() echo.HandlerFunc {
	return func(c echo.Context) error {

		log.Info("Test endpoint called seeds")

		seeds, err := tools.ReadAllSeeds()
		if err != nil {
			log.Errorf("Failed to read seeds: %v", err)
		}

		log.Info("Seeds LEN: ", len(seeds))

		return c.JSON(http.StatusOK, seeds)
	}
}
