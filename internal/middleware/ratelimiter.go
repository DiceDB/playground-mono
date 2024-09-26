package middleware

import (
	"github.com/labstack/echo/v4"
)

func RateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// implement rate limiting logic here
		return next(c)
	}
}
