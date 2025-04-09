package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Francesco99975/finexo/cmd/boot"
	"github.com/Francesco99975/finexo/internal/api"

	"github.com/Francesco99975/finexo/internal/controllers"
	"github.com/Francesco99975/finexo/internal/middlewares"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func createRouter(ctx context.Context) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Gzip())
	e.GET("/healthcheck", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		return c.JSON(http.StatusOK, "OK")
	})

	e.Static("/assets", "./static")

	web := e.Group("")

	if boot.Environment.GoEnv == "development" {
		e.Logger.SetLevel(log.DEBUG)
		log.SetLevel(log.DEBUG)
		web.Use(middlewares.SecurityHeadersDev())
	}

	if boot.Environment.GoEnv == "production" {
		e.Logger.SetLevel(log.INFO)
		log.SetLevel(log.INFO)
		web.Use(middlewares.SecurityHeaders())
	}

	web.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:_csrf,header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieHTTPOnly: true,
		Skipper: func(c echo.Context) bool {
			// Skip CSRF for the /webhook route
			return c.Path() == "/webhook"

		},
	}))

	web.GET("/", controllers.Index())
	web.GET("/req", controllers.Requests())
	web.POST("/discover", controllers.TrySeed())
	web.GET("/about", controllers.About())
	web.GET("/search", controllers.SearchHtmlSecurities())
	web.GET("/select/:tp/:id", controllers.Select())
	web.POST("/calculate", controllers.CalculateCompound())
	web.GET("/pdf/:results", controllers.DownloadPDF())
	web.GET("/csv/:results", controllers.DownloadCSV())

	apigrp := e.Group("/api")

	apiv1 := apigrp.Group("/v1")
	apiv1.GET("/search", api.SearchSecurities())
	apiv1.GET("/stocks", api.GetStocks())
	apiv1.GET("/stock/:id", api.GetStock())

	apiv1.GET("/etfs", api.GetETFs())
	apiv1.GET("/etf/:id", api.GetETF())

	apiv1.GET("/reits", api.GetREITs())
	apiv1.GET("/reit/:id", api.GetREIT())

	apiv1.GET("/test/:seed", api.Test())
	apiv1.GET("/test/seeds", api.TestSeeds())
	apiv1.GET("/test/scrape/:load", api.TestScrape())

	e.HTTPErrorHandler = serverErrorHandler

	return e
}

func serverErrorHandler(err error, c echo.Context) {
	// Default to internal server error (500)
	code := http.StatusInternalServerError
	var message any = "An unexpected error occurred"

	// Check if it's an echo.HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}

	// Check the Accept header to decide the response format
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		// Respond with JSON if the client prefers JSON
		_ = c.JSON(code, map[string]interface{}{
			"error":   true,
			"message": message,
			"status":  code,
		})
	} else {
		// Prepare data for rendering the error page (HTML)
		data := models.GetDefaultSite("Error")

		// Buffer to hold the HTML content (in case of HTML response)
		buf := bytes.NewBuffer(nil)

		// Render based on the status code

		_ = views.Error(data, fmt.Sprintf("%d", code), err).Render(context.Background(), buf)

		// Respond with HTML (default) if the client prefers HTML
		_ = c.Blob(code, "text/html; charset=utf-8", buf.Bytes())
	}
}
