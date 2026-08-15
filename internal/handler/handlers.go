package handler

import (
	"StartRomagnaAPI/config"
	"net/http"
)

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	message := "API is healthy and running on port " + config.PORT

	w.Write([]byte(message))
}
