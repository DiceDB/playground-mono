package middleware

import (
	"net/http"
	"server/config"
)

func enableCors(w http.ResponseWriter, r *http.Request) {
	configValue := config.LoadConfig()
	allAllowedOrigins := configValue.AllowedOrigins
	origin := r.Header.Get("Origin")
	allowed := false
	for _, allowedOrigin := range allAllowedOrigins {
		if origin == allowedOrigin || allowedOrigin == "*" || origin == "" {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "CORS: origin not allowed", http.StatusForbidden)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
