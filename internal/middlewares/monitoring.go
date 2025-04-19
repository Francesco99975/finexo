package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define Prometheus metrics
var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 3, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// Example custom metric: Database query latency
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"query_type"},
	)

	// Example custom metric: Business event counter
	businessEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_events_total",
			Help: "Total number of specific business events",
		},
		[]string{"event_type"},
	)
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
			httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
			httpRequestDuration.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Observe(duration)

			return err
		}
	}
}

// RecordDBQueryLatency is a helper to record database query latency
func RecordDBQueryLatency(queryType string, start time.Time) {
	duration := time.Since(start).Seconds()
	dbQueryDuration.WithLabelValues(queryType).Observe(duration)
}

// RecordBusinessEvent is a helper to record a business event
func RecordBusinessEvent(eventType string) {
	businessEventsTotal.WithLabelValues(eventType).Inc()
}

// SetupMonitoringRoutes configures the /metrics endpoint
func SetupMonitoringRoutes(e *echo.Echo) {
	// Expose Prometheus metrics endpoint
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}

// Example usage in a handler with custom metrics
func ExampleHandler(c echo.Context) error {
	// Simulate a database query
	start := time.Now()
	// ... perform database query, e.g., SELECT * FROM users
	RecordDBQueryLatency("select_users", start)

	// Simulate a business event, e.g., user login
	RecordBusinessEvent("user_login")

	// Respond based on Accept header
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
	}
	return c.HTML(http.StatusOK, "<h1>Success</h1>")
}
