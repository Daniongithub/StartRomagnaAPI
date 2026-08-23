package main

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/gtfs"
	"StartRomagnaAPI/internal/handler"
	"StartRomagnaAPI/internal/repository"
	"StartRomagnaAPI/internal/scheduler"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.LoadConf()
	repository.InitContent()

	if repository.IS_PRIMARY {
		//Operazioni per DB in modalità "primary" (non read only):

		//Viene eseguito comunque al primo avvio del programma
		gtfs.UpdateStatic()

		s, err := scheduler.InitScheduler()
		if err != nil {
			log.Fatal(err)
		}
		defer s.Shutdown()
	} else {
		println("INFO: Scheduler non avviato: rilevato DB read-only")
	}

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
	mux.HandleFunc("GET /trips", handler.TripsHandler)
	//mux.HandleFunc("GET /realtime", handler.RTHandler)

	//Listen on port and start API
	fmt.Println("Server started on port " + config.PORT)
	log.Print(http.ListenAndServe(config.PORT, mux))
}
