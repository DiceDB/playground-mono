package api

import (
	"encoding/json"
	"net/http"

	"server/internal/db"
	"server/pkg/util"
)

func JSONResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", HealthCheck)
	mux.HandleFunc("/cli/{cmd}", cliHandler)
	mux.HandleFunc("/search", searchHandler)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, r, http.StatusOK, map[string]string{"message": "Server is running"})
}

func cliHandler(w http.ResponseWriter, r *http.Request) {
	diceCmds, err :=  helpers.ParseHTTPRequest(r)
	if err!=nil{
		http.Error(w, "Error parsing HTTP request", http.StatusBadRequest)
		return
	}
	resp := db.ExecuteCommand(diceCmds)
	JSONResponse(w, r, http.StatusOK, resp)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, r, http.StatusOK, map[string]string{"message": "Search results"})
}
