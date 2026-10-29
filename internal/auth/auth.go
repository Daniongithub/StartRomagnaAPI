package auth

import (
	"io"
	"net/http"
	"time"

	"startromagnaapi/config"
)

// Client bounds network requests so a stalled feed does not occupy a job forever.
var Client = &http.Client{Timeout: 2 * time.Minute}

// BasicAuth creates a request with the upstream credentials.
func BasicAuth(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(config.WEB_AUTH_USER, config.WEB_AUTH_PASSWORD)

	return req, nil
}
