package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// this registers all API routes and middleware.
func RegisterRoutes(e *echo.Echo) {
	e.GET("/health", healthCheck)
	e.POST("/cli:command", cliHandler)
	e.GET("/search", searchHandler)
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Server is running"})
}
