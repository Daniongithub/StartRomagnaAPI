package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"startromagnaapi/config"
	"startromagnaapi/internal/cors"
	"startromagnaapi/internal/handler"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/scheduler"
	"startromagnaapi/internal/sse"

	"github.com/coreos/go-oidc/v3/oidc"
)

func main() {
	config.LoadConf()
	repository.InitStatic()
	repository.InitRT()
	repository.InitMezzi()
	repository.InitOrchestrator()
	// Initialize server-sent events for frontend synchronization.
	sse.InitHub()

	if config.IS_PRIMARY.Load() {
		// Start background updates only on the primary database.

		if err := scheduler.Start(); err != nil {
			log.Fatalf("Failed to start scheduler: %v", err)
		}
		defer scheduler.Stop()
	} else {
		log.Print("Scheduler not started: Primary database is unavailable or read-only")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthcheckHandler)

	mux.HandleFunc("GET /rss/feed", handler.RSSFeedHandler)
	mux.HandleFunc("GET /routeinfo", handler.RouteinfoHandler)

	mux.HandleFunc("GET /events", sse.SSE_HUB.SSEHandler)

	mux.HandleFunc("GET /arrivals/{stopcode}", handler.ArrivalsHandler)
	mux.HandleFunc("GET /busesinservice", handler.BusesinserviceHandler)
	mux.HandleFunc("GET /nextstops/{tripid}", handler.NextstopsHandler)
	mux.HandleFunc("GET /timetable/{basin}/{routeid}/{directionid}", handler.TimetableHandler)
	mux.HandleFunc("GET /vehiclepositions", handler.VehiclepositionsHandler)
	mux.HandleFunc("GET /vehiclepositions/{basin}", handler.VehiclepositionsBasinHandler)
	mux.HandleFunc("GET /vehicleposition/{vehicleid}", handler.VehiclepositionIDHandler)
	mux.HandleFunc("GET /shape/{basin}/{shapeId}", handler.ShapePointsHandler)
	mux.HandleFunc("GET /vehicleinfo/{vehicle}", handler.VehicleinfoHandler)
	mux.HandleFunc("GET /stopsinfo/{basin}/{stopcode}", handler.StopsinfoHandler)
	mux.HandleFunc("GET /announcements", handler.AnnouncementsHandler)

	mux.HandleFunc("GET /static/info", handler.StaticInfoHandler)
	mux.HandleFunc("GET /static/trips/{basin}", handler.TripsBasinHandler)
	mux.HandleFunc("GET /static/routes/{basin}", handler.RoutesBasinHandler)
	mux.HandleFunc("GET /static/shapes/{basin}", handler.ShapesBasinHandler)
	mux.HandleFunc("GET /static/stops/{basin}", handler.StopsBasinHandler)
	mux.HandleFunc("POST /admin/promote", handler.PromoteHandler)
	mux.HandleFunc("POST /admin/demote", handler.DemoteHandler)
	mux.HandleFunc("POST /admin/halt", handler.HaltHandler)
	mux.HandleFunc("POST /admin/resync", handler.ResyncHandler)

	
	if config.ENABLE_AUTH {
		ctx := context.Background()

		provider, err := oidc.NewProvider(ctx, config.OIDC_ISSUER_URL)
		if err != nil {
			log.Fatal("impossibile contattare Authentik per il discovery OIDC:", err)
		}

		verifier := provider.Verifier(&oidc.Config{ClientID: config.OIDC_CLIENT_ID})
		authMiddleware := handler.RequireAuth(verifier, config.OIDC_ALLOWED_GROUPS)

		mux.Handle("PUT /announcements/update", authMiddleware(http.HandlerFunc(handler.UpdateAnnouncementHandler)))
		mux.Handle("DELETE /announcements/empty", authMiddleware(http.HandlerFunc(handler.EmptyAnnouncementsHandler)))
	}

	// Listen on the configured port and start the API.
	log.Printf("Starting HTTP server on %s", config.PORT)
	server := &http.Server{Addr: config.PORT, Handler: cors.Middleware(mux), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("HTTP server stopped: %v", server.ListenAndServe())
}
