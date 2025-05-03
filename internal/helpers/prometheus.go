package helpers

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

func init() {
	// Register custom metrics
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, dbQueryDuration, businessEventsTotal)
}

func IncreaseHTTPRequestCount(method, path string, status int) {
	httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
}

func RecordHTTPRequestDuration(method, path string, status int, duration float64) {
	httpRequestDuration.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Observe(duration)
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

// Example usage in a handler with custom metrics
// func ExampleHandler(c echo.Context) error {
// 	// Simulate a database query
// 	start := time.Now()
// 	// ... perform database query, e.g., SELECT * FROM users
// 	RecordDBQueryLatency("select_users", start)

// 	// Simulate a business event, e.g., user login
// 	RecordBusinessEvent("user_login")

// 	// Respond based on Accept header
// 	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
// 		return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
// 	}
// 	return c.HTML(http.StatusOK, "<h1>Success</h1>")
// }
