package api

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all API routes and middleware.
func RegisterRoutes(e *echo.Echo) {
	e.POST("/cli:command", cliHandler)
	e.GET("/search", searchHandler)
}

func ParseCommand(command string) (string, error) {
	parsedCommand := strings.TrimPrefix(command, "/")
	if parsedCommand == "" {
		return "", errors.New("command parser error | empty command")
	}
	return parsedCommand, nil
}
