package handler

import (
	"StartRomagnaAPI/config"
	"net/http"
	"strings"
)

func AddCORS(w http.ResponseWriter, r *http.Request) {

	origin := r.Header.Get("Origin")

	if origin != "" && isAllowedOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}
}

func isAllowedOrigin(origin string) bool {
	for _, allowedOrigin := range config.ALLOWED_ORIGINS {
		if strings.TrimSpace(allowedOrigin) == origin {
			return true
		}
	}

	//Localhost with any port
	return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
}
