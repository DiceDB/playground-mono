package api

import (
	"net/http"

	"server/internal/db"

	"github.com/labstack/echo/v4"
)

type CommandRequest struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
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
		if req.Key == "" || req.Value == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Key and Value are required"})
		}
		err := db.SetKey(req.Key, req.Value)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set key"})
		}
		return c.JSON(http.StatusOK, map[string]string{"result": "OK"})

	case "get":
		if req.Key == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Key is required"})
		}
		result, err := db.GetKey(req.Key)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get key"})
		}
		return c.JSON(http.StatusOK, map[string]string{"key": req.Key, "value": result})

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid command"})
	}
}
