package middlewares

import (
	"net/http"

	"slices"

	"github.com/Francesco99975/finexo/cmd/boot"
	"github.com/labstack/echo/v4"
)

func QueryParamMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if boot.Environment.GoEnv == "production" {
				proxySecret := c.Request().Header.Get("X-RapidAPI-Proxy-Secret")
				if proxySecret != boot.Environment.RapidApiSecret { // Use your actual secret here
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid request source")
				}
				// If secret matches, proceed with the request
			}

			// Extract subscription plan from header
			subscription := c.Request().Header.Get("X-RapidAPI-Subscription")
			if subscription == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing subscription information")
			}

			// Get plan limits
			plan, exists := boot.PlanConfigs[subscription]
			if !exists {
				return echo.NewHTTPError(http.StatusForbidden, "Invalid subscription plan")
			}

			// Get query parameters
			queryParams := c.QueryParams()

			// Check maximum parameter limit
			if plan.MaxParams != -1 && len(queryParams) > plan.MaxParams {
				return echo.NewHTTPError(http.StatusForbidden, "Exceeded maximum allowed query parameters")
			}

			// Check allowed parameters
			if plan.AllowedParams != nil {
				for param := range queryParams {
					if !contains(plan.AllowedParams, param) {
						return echo.NewHTTPError(http.StatusForbidden, "Unauthorized query parameter: "+param)
					}
				}
			}

			// Proceed to the handler if all checks pass
			return next(c)
		}
	}
}

// Helper function to check if a string is in a slice
func contains(arr []string, str string) bool {
	return slices.Contains(arr, str)
}
