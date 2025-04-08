package controllers

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views"
	"github.com/labstack/echo/v4"
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
