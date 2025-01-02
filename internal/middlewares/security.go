package middlewares

import (
	"fmt"

	"github.com/Francesco99975/finexo/cmd/boot"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/labstack/echo/v4"
)

func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			nonce, err := helpers.GenerateNonce()
			if err != nil {
				return err
			}

			// Store nonce in the request context so it can be accessed in your templates
			c.Set("nonce", nonce)

			secureWebsocketUri := fmt.Sprintf("wss://%s", boot.Environment.Host) // Allow Secure Websocket connections

			// Set security headers
			c.Response().Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'nonce-"+nonce+"' 'strict-dynamic'; "+
					"connect-src 'self' '"+secureWebsocketUri+"'; "+
					"style-src 'self'; "+
					"frame-src")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "no-referrer")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(self), microphone=()")
			return next(c)
		}
	}
}

func SecurityHeadersDev() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			nonce, err := helpers.GenerateNonce()
			if err != nil {
				return err
			}

			// Store nonce in the request context so it can be accessed in your templates
			c.Set("nonce", nonce)

			// Set security headers
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'nonce-"+nonce+"' 'strict-dynamic' 'unsafe-eval'; connect-src 'self'; style-src 'self' 'unsafe-inline'; frame-src")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "SAMEORIGIN") // Allow iframes for easier testing
			// No Strict-Transport-Security for local development
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "no-referrer-when-downgrade") // Less strict for development
			return next(c)
		}
	}
}
