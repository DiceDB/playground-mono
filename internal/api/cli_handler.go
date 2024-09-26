package api

import (
	"net/http"

	"server/internal/db"

	"github.com/labstack/echo/v4"
)

type CommandRequest struct {
	Key   string   `json:"key"`
	Value string   `json:"value,omitempty"`
	Keys  []string `json:"keys,omitempty"`
}

func handleSet(req CommandRequest) error {
	if req.Key == "" || req.Value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Key and Value are required")
	}
	err := db.SetKey(req.Key, req.Value)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set key")
	}
	return echo.NewHTTPError(http.StatusOK, map[string]string{"result": "OK"})
}

func handleGet(req CommandRequest) error {
	if req.Key == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Key is required")
	}
	result, err := db.GetKey(req.Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get key")
	}
	return echo.NewHTTPError(http.StatusOK, map[string]string{"key": req.Key, "value": result})
}

func handleDel(req CommandRequest) error {
	if len(req.Keys) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one key is required for deletion")
	}
	err := db.DeleteKeys(req.Keys)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete keys")
	}
	return echo.NewHTTPError(http.StatusOK, map[string]string{"result": "OK"})
}

func cliHandler(c echo.Context) error {
	command, err := ParseCommand(c.Param("command"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "command parser error | empty command"})
	}

	var req CommandRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	switch command {
	case "set":
		return handleSet(req)

	case "get":
		return handleGet(req)

	case "del":
		return handleDel(req)

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid command"})
	}
}
