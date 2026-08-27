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

	if config.IS_PRIMARY {
		//Operazioni per DB in modalità "primary" (non read only):

		//Viene eseguito comunque al primo avvio del programma
		go gtfs.UpdateStatic()

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
	mux.HandleFunc("GET /feed", handler.FeedHandler)
	mux.HandleFunc("GET /static/trips", handler.TripsHandler)
	mux.HandleFunc("GET /static/trips/{basin}", handler.TripsBasinHandler)
	mux.HandleFunc("GET /static/calendar_dates", handler.CalDatesHandler)
	mux.HandleFunc("GET /static/calendar_dates/{basin}", handler.CalDatesBasinHandler)
	mux.HandleFunc("GET /static/routes", handler.RoutesHandler)
	mux.HandleFunc("GET /static/routes/{basin}", handler.RoutesBasinHandler)
	mux.HandleFunc("GET /static/shapes", handler.ShapesHandler)
	mux.HandleFunc("GET /static/shapes/{basin}", handler.RoutesBasinHandler)
	mux.HandleFunc("GET /static/stop_times", handler.StopTimesHandler)
	mux.HandleFunc("GET /static/stop_times/{basin}", handler.StopTimesBasinHandler)
	mux.HandleFunc("GET /static/stops", handler.StopsHandler)
	mux.HandleFunc("GET /static/stops/{basin}", handler.StopsBasinHandler)
	//mux.HandleFunc("GET /realtime", handler.RTHandler)

	//Listen on port and start API
	fmt.Println("Server started on port " + config.PORT)
	log.Print(http.ListenAndServe(config.PORT, mux))
}
