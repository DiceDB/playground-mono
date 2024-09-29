package handlers

import (
	"net/http"
	"server/internal/service"
	"server/pkg/common"
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
	}

	var result any
	switch command.Cmd {

	case OpGET:
		result, err = h.service.Get(r.Context(), &service.GetRequest{Key: command.Args.Key})

	case OpSET:
		result, err = h.service.Set(r.Context(), &service.SetRequest{Key: command.Args.Key, Value: command.Args.Value})

	case OpDEL:
		var keys []string
		if command.Args.Key != "" {
			keys = []string{command.Args.Key}
		} else if len(command.Args.Keys) > 0 {
			keys = command.Args.Keys
		} else {
			common.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "at least one key is required"})
			return
		}
		result, err = h.service.Delete(r.Context(), &service.DeleteRequest{Keys: keys})
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
