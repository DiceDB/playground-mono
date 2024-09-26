package api

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", HealthCheck)
	mux.HandleFunc("/cli", cliHandler)
	mux.HandleFunc("/search", searchHandler)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, r, http.StatusOK, map[string]string{"message": "Server is running"})
}

func cliHandler(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, r, http.StatusOK, map[string]string{"message": "cli handler"})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, r, http.StatusOK, map[string]string{"message": "Search results"})
}
