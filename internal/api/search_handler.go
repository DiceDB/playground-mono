package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func searchHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Search results"})
}
