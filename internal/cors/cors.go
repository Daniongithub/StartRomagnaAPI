package cors

import (
	"net/http"
	"net/url"
	"strings"

	"startromagnaapi/config"
)

// Middleware adds CORS headers before the mux handles any response, including errors.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Add(w, r)
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" &&
			r.Header.Get("Origin") != "" && isAllowedOrigin(r.Header.Get("Origin")) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Add(w http.ResponseWriter, r *http.Request) {

	origin := r.Header.Get("Origin")
	w.Header().Add("Vary", "Origin")

	if origin != "" && isAllowedOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	}
}

func isAllowedOrigin(origin string) bool {
	for _, allowedOrigin := range config.ALLOWED_ORIGINS {
		if strings.TrimSpace(allowedOrigin) == origin {
			return true
		}
	}

	// Allow localhost origins on any port.
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "http" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	return u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"
}
