package middleware

import (
	"net/http"
	"server/config"
)

func enableCors(w http.ResponseWriter, origin string) {
	configValue := config.LoadConfig()
	allAllowedOrigins := configValue.AllowedOrigins
	allowed := false
	for _, allowedOrigin := range allAllowedOrigins {
		if origin == allowedOrigin || allowedOrigin == "*" {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "CORS: origin not allowed", http.StatusForbidden)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
}
