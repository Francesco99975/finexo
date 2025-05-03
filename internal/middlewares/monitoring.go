package middlewares

import (
	"strings"
	"time"

	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/labstack/echo/v4"
)

// MonitoringMiddleware tracks request metrics and exposes them for Prometheus
func MonitoringMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			path := c.Path()
			method := c.Request().Method

			// Proceed with the request
			err := next(c)

			// Calculate duration
			duration := time.Since(start).Seconds()
			status := c.Response().Status

			// Sanitize path for metrics (e.g., convert dynamic routes like /user/:id to /user/{id})
			if strings.Contains(path, ":") {
				path = strings.ReplaceAll(path, ":", "{") + "}"
			}

			// Record metrics
			helpers.IncreaseHTTPRequestCount(method, path, status)
			helpers.RecordHTTPRequestDuration(method, path, status, duration)

			return err
		}
	}
}
