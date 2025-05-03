package controllers

import (
	"net/http"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/Francesco99975/finexo/views"
	"github.com/Francesco99975/finexo/views/components"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Requests() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := models.GetDefaultSite("Requests")

		csrfToken := c.Get("csrf").(string)
		nonce := c.Get("nonce").(string)

		html, err := helpers.RenderHTML(views.Requests(data, csrfToken, nonce))

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse page about")
		}

		return c.Blob(200, "text/html; charset=utf-8", html)
	}
}

func TrySeed() echo.HandlerFunc {
	return func(c echo.Context) error {
		type SeedFormData struct {
			Ticker   string `form:"ticker"`
			Exchange string `form:"exchange"`
		}
		var input SeedFormData

		if err := c.Bind(&input); err != nil {
			log.Errorf("invalid form data for requests: %w", err)

			html := helpers.MustRenderHTML(components.ErrorMsg("Invalid form data"))

			return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
		}

		if input.Ticker == "" || input.Exchange == "" {
			log.Error("missing ticker or exchange")

			html := helpers.MustRenderHTML(components.ErrorMsg("Missing ticker or exchange"))

			return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
		}

		start := time.Now()

		if models.SecurityExists(database.DB, input.Ticker, input.Exchange) {
			log.Warn("security already exists")
			html := helpers.MustRenderHTML(components.WarnMsg("Security Already Exists"))

			return c.Blob(http.StatusAlreadyReported, "text/html; charset=utf-8", html)
		}
		manager := models.NewBrowserManager(500)

		d, err := tools.NewDiscoverer()
		if err != nil {
			log.Warnf("failed to create discoverer: %v", err)
		}

		err = tools.Scrape(input.Ticker, &input.Exchange, manager, nil, nil, d)
		if err != nil {
			html := helpers.MustRenderHTML(components.ErrorMsg("Security could not be scraped"))

			return c.Blob(http.StatusBadRequest, "text/html; charset=utf-8", html)
		}
		helpers.RecordDBQueryLatency("request_security", start)
		helpers.RecordBusinessEvent("request_security")

		html := helpers.MustRenderHTML(components.SuccessMsg("Security discovered successfully!"))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}
