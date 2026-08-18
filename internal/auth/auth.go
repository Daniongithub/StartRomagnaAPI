package auth

import (
	"StartRomagnaAPI/config"
	"io"
	"net/http"
)

// BasicAuth restituisce un http.Client configurato per autenticarsi
// tramite Basic Auth, anche dopo eventuali redirect.
func BasicAuth(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(config.WEB_AUTH_USER, config.WEB_AUTH_PASSWORD)

	return req, nil
}
