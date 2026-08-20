package main

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/handler"
	"StartRomagnaAPI/internal/repository"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.LoadConf()
	repository.InitContent()

	mux := http.NewServeMux()
	//Redirect root to healthcheck
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/health", http.StatusFound)
			return
		}

		http.NotFound(w, r)
	})

	mux.HandleFunc("GET /health", handler.HealthcheckHandler)
	mux.HandleFunc("GET /test", handler.GTFSHandler)
	mux.HandleFunc("GET /realtime", handler.TripsHandler)

	//Listen on port and start API
	fmt.Println("Server started on port " + config.PORT)
	log.Print(http.ListenAndServe(config.PORT, mux))
}
