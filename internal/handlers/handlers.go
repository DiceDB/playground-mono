package handlers

import (
	"net/http"

	"github.com/DiceDB/playground-mono/internal/service"
	"github.com/DiceDB/playground-mono/pkg/common"
)

const (
	OpGET = "get"
	OpSET = "set"
	OpDEL = "delete"
)

type Handler interface {
	CLIHandler(w http.ResponseWriter, r *http.Request)
	Health(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service service.Service
}

func NewHandler(service service.Service) Handler {
	return &handler{service: service}
}

func (h *handler) CLIHandler(w http.ResponseWriter, r *http.Request) {
	command, err := common.ParseHTTPRequest(r)
	if err != nil {
		common.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	var result any
	switch command.Cmd {
	case OpGET:
		if len(command.Args) < 1 {
			common.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "key is required for GET operation"})
			return
		}
		result, err = h.service.Get(r.Context(), &service.GetRequest{Key: command.Args[0]})

	case OpSET:
		if len(command.Args) < 2 {
			common.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "key and value are required for SET operation"})
			return
		}
		result, err = h.service.Set(r.Context(), &service.SetRequest{Key: command.Args[0], Value: command.Args[1]})

	case OpDEL:
		if len(command.Args) < 1 {
			common.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "at least one key is required for DELETE operation"})
			return
		}
		result, err = h.service.Delete(r.Context(), &service.DeleteRequest{Keys: command.Args})

	default:
		common.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "unknown command"})
		return
	}

	if err != nil {
		common.JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	common.JSONResponse(w, http.StatusOK, result)
}

func (h *handler) Health(w http.ResponseWriter, r *http.Request) {
	// TODO: Call service layer for Health check
	common.JSONResponse(w, http.StatusOK, map[string]string{"message": "Server is running"})
}

func (h *handler) Search(w http.ResponseWriter, r *http.Request) {
	// TODO: Call service layer for search over keys or whatever
	common.JSONResponse(w, http.StatusOK, map[string]string{"message": "Results..."})
}
