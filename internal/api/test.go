package api

import (
	"net/http"
	"strconv"

	"github.com/Francesco99975/finexo/cmd/boot"
	"github.com/Francesco99975/finexo/internal/models"
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

		manager := models.NewBrowserManager(500)

		d, err := tools.NewDiscoverer()
		if err != nil {
			log.Errorf("Failed to create discoverer: %v", err)
		}

		err = tools.Scrape(seed, nil, manager, nil, nil, d)
		if err != nil {
			log.Errorf("Failed to scrape: %v", err)
		}

		return c.JSON(http.StatusOK, "OK")
	}
}

func TestScrape() echo.HandlerFunc {
	return func(c echo.Context) error {
		var load int
		var err error
		loadParam := c.Param("load")
		if loadParam == "" {
			load = 30
		} else {
			load, err = strconv.Atoi(loadParam)
			if err != nil {
				return c.JSON(http.StatusBadRequest, "Invalid load parameter")
			}
		}

		err = boot.SeedDatabase(load, "")
		if err != nil {
			log.Errorf("Failed to seed database: %v", err)
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
