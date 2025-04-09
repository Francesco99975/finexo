package controllers

import (
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views"
	"github.com/labstack/echo/v4"
)

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := models.GetDefaultSite("Home")

		csrfToken := c.Get("csrf").(string)
		nonce := c.Get("nonce").(string)

		html := helpers.MustRenderHTML(views.Index(data, csrfToken, nonce))

		return c.Blob(200, "text/html; charset=utf-8", html)
	}
}
