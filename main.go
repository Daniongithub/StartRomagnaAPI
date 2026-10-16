package main

import (
	//"context"
	"fmt"
	"log"
	"net/http"
	"startromagnaapi/config"
	"startromagnaapi/internal/handler"
	"startromagnaapi/internal/scheduler"

	//"startromagnaapi/internal/panels/announcements"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/sse"
)

func main() {
	config.LoadConf()
	repository.InitStatic()
	repository.InitRT()
	repository.InitMezzi()
	//Initializes server send events for fontend sync
	sse.InitHub()

	if config.IS_PRIMARY.Load() {
		//Operazioni per DB in modalità "primary" (non read only):

		var err error
		config.SCHEDULER, err = scheduler.InitScheduler()
		if err != nil {
			log.Fatal(err)
		}
		defer config.SCHEDULER.Shutdown()
	} else {
		println("INFO: Scheduler non avviato: rilevato DB read-only")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthcheckHandler)

	mux.HandleFunc("GET /rss/feed", handler.RSSFeedHandler)
	mux.HandleFunc("GET /routeproblems", handler.RouteProblemsHandler)

	mux.HandleFunc("/events", sse.SSE_HUB.SSEHandler)

	mux.HandleFunc("GET /arrivals/{stopcode}", handler.ArrivalsHandler)
	mux.HandleFunc("GET /busesinservice", handler.BusesinserviceHandler)
	mux.HandleFunc("GET /nextstops/{tripid}", handler.NextstopsHandler)
	mux.HandleFunc("GET /vehiclepositions", handler.VehiclepositionsHandler)
	mux.HandleFunc("GET /vehiclepositions/{basin}", handler.VehiclepositionsBasinHandler)
	mux.HandleFunc("GET /vehicleposition/{vehicleid}", handler.VehiclepositionIDHandler)
	mux.HandleFunc("GET /shape/{basin}/{shapeId}", handler.ShapePointsHandler)
	mux.HandleFunc("GET /vehicleinfo/{vehicle}", handler.VehicleinfoHandler)
	mux.HandleFunc("GET /stopsinfo/{basin}/{stopcode}", handler.StopsinfoHandler)
	mux.HandleFunc("GET /announcements", handler.AnnouncementsHandler)

	//mux.HandleFunc("GET /timetable/{routeid}", handler.TimetableHandler)

	mux.HandleFunc("GET /static/info", handler.StaticInfoHandler)
	mux.HandleFunc("GET /static/trips/{basin}", handler.TripsBasinHandler)
	mux.HandleFunc("GET /static/routes/{basin}", handler.RoutesBasinHandler)
	mux.HandleFunc("GET /static/shapes/{basin}", handler.ShapesBasinHandler)
	mux.HandleFunc("GET /static/stops/{basin}", handler.StopsBasinHandler)
	mux.HandleFunc("POST /admin/promote", handler.PromoteHandler)
	mux.HandleFunc("POST /admin/demote", handler.DemoteHandler)
	mux.HandleFunc("POST /admin/halt", handler.HaltHandler)
	mux.HandleFunc("POST /admin/resync", handler.ResyncHandler)

	/*
		if config.ENABLE_AUTH {
			panel, err := announcements.New(context.Background())
			if err != nil {
				log.Fatalf("inizializzazione pannello annunci fallita: %v", err)
			}
			if err := panel.RegisterAPI("GET", "/api/announcements", http.HandlerFunc(handler.AnnouncementsHandler)); err != nil {
				log.Fatalf("registrazione API pannello annunci fallita: %v", err)
			}
			panel.Mount(mux)
		}*/

	//Listen on port and start API
	fmt.Println("Server started on port " + config.PORT)
	log.Print(http.ListenAndServe(config.PORT, mux))
}
