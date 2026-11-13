package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"startromagnaapi/config"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/service"
	"strconv"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// GET /health
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	results, err := repository.GetDBStatus(repository.DB_STATIC.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /rss/feed
func RSSFeedHandler(w http.ResponseWriter, r *http.Request) {
	results, err := service.GetRSSFeed(config.FEED_URL)

	if err != nil {
		if errors.Is(err, service.ErrFeed) {
			http.Error(w, "Rss feed error", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(results)
}

// GET /routeinfo
func RouteinfoHandler(w http.ResponseWriter, r *http.Request) {
	results := static.GetRouteInfos()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /arrivals/{stopcode}
func ArrivalsHandler(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("stopcode")

	results, err := service.ProcessArrivals(code)
	if err != nil {
		if errors.Is(err, service.NoArrivals) {
			http.Error(w, "No arrivals", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /busesinservice
func BusesinserviceHandler(w http.ResponseWriter, r *http.Request) {
	results, err := service.ProcessBusesInService()
	if err != nil {
		if errors.Is(err, service.NoBusesInService) {
			http.Error(w, "No buses in service at the moment", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /nextstops/{tripid}
func NextstopsHandler(w http.ResponseWriter, r *http.Request) {
	tripid := r.PathValue("tripid")

	results := service.ProcessNextstops(tripid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /timetable/{basin}/{routeid}/{directionid}
func TimetableHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")
	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}
	routeid := r.PathValue("routeid")
	directionid := r.PathValue("directionid")
	directionInt, err := strconv.Atoi(directionid)
	if err != nil || directionInt < 0 || directionInt > 1 {
		http.Error(w, "Bad request on direction parameter, can be only 0 or 1.", http.StatusBadRequest)
		return
	}
	results := service.ProcessTimetable(basin, routeid, directionid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /vehiclepositions
func VehiclepositionsHandler(w http.ResponseWriter, r *http.Request) {
	results := service.ProcessVehiclePositions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /vehiclepos/{basin}
func VehiclepositionsBasinHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := service.ProcessVehiclePositionsBasin(basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /vehicleposition/{vehicleid}
func VehiclepositionIDHandler(w http.ResponseWriter, r *http.Request) {
	vehicleId := r.PathValue("vehicleid")

	results := service.ProcessVehiclePositionID(vehicleId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /shape/{basin}/{shapeId}
func ShapePointsHandler(w http.ResponseWriter, r *http.Request) {
	shape_id := r.PathValue("shapeId")

	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := service.ProcessShape(basin, shape_id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /shapeslist/{basin}/{routeId}
func ShapesListHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")
	routeid := r.PathValue("routeId")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetHeadsignsByRouteID(basin, routeid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /vehicleinfo/{vehicle}
func VehicleinfoHandler(w http.ResponseWriter, r *http.Request) {
	vehicleId := r.PathValue("vehicle")

	results, err := service.ProcessVehicleInfo(vehicleId)
	if err != nil {
		if errors.Is(err, service.ErrVehicleNotFound) {
			http.Error(w, "Vehicle not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /stopsinfo/{basin}/{stopcode}
func StopsinfoHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")
	stopCode := r.PathValue("stopcode")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}
	results := service.ProcessStopsInfo(stopCode, basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /announcements
func AnnouncementsHandler(w http.ResponseWriter, r *http.Request) {
	announcements := static.GetAnnouncements()
	if len(announcements) == 0 {
		http.Error(w, "No announcements", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

// =======================
// = RAW GTFS ENDPOINTS  =
// =======================

// GET /static/info
func StaticInfoHandler(w http.ResponseWriter, r *http.Request) {

	min_dateRA, max_dateRA := static.GetCalDatesRange("RA")
	min_dateFC, max_dateFC := static.GetCalDatesRange("FC")
	min_dateRN, max_dateRN := static.GetCalDatesRange("RN")

	message := "GTFS static data of Ravenna valid from " + min_dateRA + " to " + max_dateRA + "." + "\n\n"
	message += "GTFS static data of Forlì-Cesena valid from " + min_dateFC + " to " + max_dateFC + "." + "\n\n"
	message += "GTFS static data of Rimini valid from " + min_dateRN + " to " + max_dateRN + "." + "\n\n"

	w.Write([]byte(message))
}

// GET /static/trips/{basin}
func TripsBasinHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetTripsBasin(basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /static/routes/{basin}
func RoutesBasinHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := service.ProcessRoutes(basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /static/shapes/{basin}
func ShapesBasinHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetShapesBasin(basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /static/stops/{basin}
func StopsBasinHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetStopsBasin(basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// PUT /announcements/update (AUTH REQUIRED)
func UpdateAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	//Sets headers
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType) // 415
		json.NewEncoder(w).Encode(map[string]string{"error": "Expected application/json."})
		return
	}
	var req model.Announcement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body."})
		return
	}
	defer r.Body.Close()

	//Deletes old announcement
	err := static.EmptyAnnouncements()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	log.Println(req.Expiry)
	err = static.InsertAnnouncement(req.Message, req.Expiry)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	message := "Add request OK"

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode([]byte(message))
}

// DELETE /announcements/empty (AUTH REQUIRED)
func EmptyAnnouncementsHandler(w http.ResponseWriter, r *http.Request) {
	//Sets headers
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType) // 415
		json.NewEncoder(w).Encode(map[string]string{"error": "Expected application/json."})
		return
	}

	err := static.EmptyAnnouncements()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	message := "Add request OK"

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode([]byte(message))
}

// Auth flow
func RequireAuth(verifier *oidc.IDTokenVerifier, allowedGroups []string) func(http.Handler) http.Handler {
	allowedSet := make(map[string]bool, len(allowedGroups))
	for _, g := range allowedGroups {
		allowedSet[g] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Missing bearer token."})
				return
			}
			rawToken := strings.TrimPrefix(authHeader, "Bearer ")

			idToken, err := verifier.Verify(r.Context(), rawToken)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired token."})
				return
			}

			var claims struct {
				Groups []string `json:"groups"`
			}
			if err := idToken.Claims(&claims); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Malformed token claims."})
				return
			}

			authorized := false
			for _, g := range claims.Groups {
				if allowedSet[g] {
					authorized = true
					break
				}
			}
			if !authorized {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "Insufficient group membership."})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
