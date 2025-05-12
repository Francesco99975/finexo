package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(30, 10) // 1 request per second, with a burst of 3

func RateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !limiter.Allow() {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"message": "Too many requests, please try again later.",
			})
		}
		return next(c)
	}
}

// Create a specific limiter for thr POST /discover endpoint to limit up to 3 request per day
// var discoverLimiter = rate.NewLimiter(3, 3)

// func DiscoverRateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		if !discoverLimiter.Allow() {
// 			return c.JSON(http.StatusTooManyRequests, map[string]string{
// 				"message": "Too many requests, please try again later.",
// 			})
// 		}
// 		return next(c)
// 	}
// }
