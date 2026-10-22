package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"startromagnaapi/config"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/service"
	"strconv"
	"sync/atomic"
)

const (
	FEED_URL       = "https://www.startromagna.it/infobus/feed/"
	ROUTE_INFO_URL = "https://www.startromagna.it/orari-e-percorsi-vers-2-0/"
)

// GET /health
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)

	repository.IsPrimary(repository.DB_STATIC)
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
	AddCORS(w, r)
	results, err := service.GetRSSFeed(FEED_URL)

	if err != nil {
		if errors.Is(err, service.NoArrivals) {
			http.Error(w, "rss feed error", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
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
	AddCORS(w, r)
	results := static.GetRouteInfos()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /arrivals/{stopcode}
func ArrivalsHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
	code := r.PathValue("stopcode")

	results, err := service.ProcessArrivals(code)
	if err != nil {
		if errors.Is(err, service.NoArrivals) {
			http.Error(w, "no arrivals", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /busesinservice
func BusesinserviceHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
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
	AddCORS(w, r)
	tripid := r.PathValue("tripid")

	results := service.ProcessNextstops(tripid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /timetable/{basin}/{routeid}/{directionid}
func TimetableHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
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
	AddCORS(w, r)
	results := service.ProcessVehiclePositions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /vehiclepos/{basin}
func VehiclepositionsBasinHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
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
	AddCORS(w, r)
	vehicleId := r.PathValue("vehicleid")

	results := service.ProcessVehiclePositionID(vehicleId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /shape/{basin}/{shapeId}
func ShapePointsHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
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

// GET /vehicleinfo/{vehicle}
func VehicleinfoHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
	vehicleId := r.PathValue("vehicle")

	results, err := service.ProcessVehicleInfo(vehicleId)
	if err != nil {
		if errors.Is(err, service.ErrVehicleNotFound) {
			http.Error(w, "vehicle not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /stopsinfo/{basin}/{stopcode}
func StopsinfoHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")
	stopCode := r.PathValue("stopcode")

	AddCORS(w, r)
	results := service.ProcessStopsInfo(stopCode, basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /announcements
func AnnouncementsHandler(w http.ResponseWriter, r *http.Request) {
	announcements := static.GetAnnouncements()

	AddCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

// =======================
// = RAW GTFS ENDPOINTS  =
// =======================

// GET /static/info
func StaticInfoHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)

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
	AddCORS(w, r)
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
	AddCORS(w, r)
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetRoutesBasin(basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /static/shapes/{basin}
func ShapesBasinHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
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
	AddCORS(w, r)
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetStopsBasin(basin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func arbiterAuthorized(r *http.Request) bool {
	return config.ARBITER_SECRET != "" && r.Header.Get("X-Arbiter-Secret") == config.ARBITER_SECRET
}

// POST /admin/promote
func PromoteHandler(w http.ResponseWriter, r *http.Request) {
	if !arbiterAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := repository.DB_ORCH // same db instance

	if _, err := db.Exec(`STOP SLAVE`); err != nil {
		http.Error(w, fmt.Sprintf("Stop slave failed: %v", err), http.StatusInternalServerError)
		return
	}
	query := `
		GRANT INSERT, UPDATE, DELETE ON start_gtfs_static.* TO '` + config.DB_USERNAME + `'@'%';
		GRANT INSERT, UPDATE, DELETE ON start_gtfs_rt.*     TO '` + config.DB_USERNAME + `'@'%';
		GRANT UPDATE ON ertpl_mezzi.*                       TO '` + config.DB_USERNAME + `'@'%';
		FLUSH PRIVILEGES;
	`
	if _, err := db.Exec(`
		GRANT INSERT, UPDATE, DELETE ON start_gtfs_static.* TO '` + config.DB_USERNAME + `'@'%';
		GRANT INSERT, UPDATE, DELETE ON start_gtfs_rt.*     TO '` + config.DB_USERNAME + `'@'%';
		GRANT UPDATE ON ertpl_mezzi.*                       TO '` + config.DB_USERNAME + `'@'%';
		FLUSH PRIVILEGES;
	`); err != nil {
		http.Error(w, fmt.Sprintf("Grant privileges failed: %v", err, query), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "promoted"})
}

type demoteRequest struct {
	NewMasterHost string `json:"newMasterHost"`
}

// POST /admin/demote
func DemoteHandler(w http.ResponseWriter, r *http.Request) {
	if !arbiterAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req demoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewMasterHost == "" {
		http.Error(w, "Invalid or missing newMasterHost", http.StatusBadRequest)
		return
	}

	if config.REPLICA_USERNAME == "" || config.REPLICA_PASSWORD == "" {
		http.Error(w, "Replica not configured", http.StatusInternalServerError)
		return
	}

	db := repository.DB_ORCH

	if _, err := db.Exec(`
		REVOKE INSERT, UPDATE, DELETE ON start_gtfs_static.* FROM '` + config.DB_USERNAME + `'@'%';
		REVOKE INSERT, UPDATE, DELETE ON start_gtfs_rt.*     FROM '` + config.DB_USERNAME + `'@'%';
		REVOKE UPDATE ON ertpl_mezzi.*                       FROM '` + config.DB_USERNAME + `'@'%';
		FLUSH PRIVILEGES;
	`); err != nil {
		http.Error(w, fmt.Sprintf("Revoke privileges failed: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := db.Exec("STOP SLAVE"); err != nil {
		http.Error(w, fmt.Sprintf("Stop slave failed: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := db.Exec(
		"CHANGE MASTER TO MASTER_HOST=?, MASTER_PORT=3306, MASTER_USER=?, MASTER_PASSWORD=?, MASTER_USE_GTID=slave_pos",
		req.NewMasterHost, config.REPLICA_USERNAME, config.REPLICA_PASSWORD,
	); err != nil {
		http.Error(w, fmt.Sprintf("Change master failed: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := db.Exec("START SLAVE"); err != nil {
		http.Error(w, fmt.Sprintf("Ctart slave failed: %v", err), http.StatusInternalServerError)
		return
	}
	repository.IsPrimary(repository.DB_STATIC)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "demoted", "newMaster": req.NewMasterHost})
}

// POST /admin/halt
func HaltHandler(w http.ResponseWriter, r *http.Request) {
	if !arbiterAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := repository.HaltLocal(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repository.IsPrimary(repository.DB_STATIC)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "unconfigured"})
}

type resyncRequest struct {
	MasterHost string `json:"masterHost"`
}

// POST /admin/resync
func ResyncHandler(w http.ResponseWriter, r *http.Request) {
	if !arbiterAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req resyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MasterHost == "" {
		http.Error(w, "Invalid or missing masterHost", http.StatusBadRequest)
		return
	}

	var resyncInProgress atomic.Bool
	
    if !arbiterAuthorized(r) { /* ... */ }
    if !resyncInProgress.CompareAndSwap(false, true) {
        http.Error(w, "resync già in corso", http.StatusConflict)
        return
    }
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{"status": "resync avviato"})
    go func() {
        defer resyncInProgress.Store(false)
        repository.RunResync(req.MasterHost)
    }()
}
