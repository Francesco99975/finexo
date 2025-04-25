package controllers

import (
	"net/http"

	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views"
	"github.com/labstack/echo/v4"
)

func About() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := models.GetDefaultSite("About")

		csrfToken := c.Get("csrf").(string)
		nonce := c.Get("nonce").(string)

		html := helpers.MustRenderHTML(views.About(data, csrfToken, nonce))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}
