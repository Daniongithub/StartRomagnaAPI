package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"startromagnaapi/config"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/service"
	"time"

	"github.com/mmcdole/gofeed"
)

const feedURL = "https://www.startromagna.it/infobus/feed/"

// GET /health
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)

	message := "API is healthy and running on port " + config.PORT

	w.Write([]byte(message))
}

// GET /rss/feed
func RSSFeedHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(feedURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := model.FeedResponse{
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		Items:       make([]model.FeedItem, 0, len(feed.Items)),
	}
	for _, item := range feed.Items {
		author := ""
		if len(item.Authors) > 0 && item.Authors[0] != nil {
			author = item.Authors[0].Name
		}

		var published time.Time
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}
		response.Items = append(response.Items, model.FeedItem{
			Title:           item.Title,
			Description:     item.Description,
			Content:         item.Content,
			Link:            item.Link,
			Author:          author,
			Published:       item.Published,
			PublishedParsed: published,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)
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
			http.Error(w, "no buses in service at the moment", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /activevehicles
func ActivevehiclesHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
	results := realtime.GetVehicles()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /linelist/{basin}
func LinelistHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetRouteName(basin)

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

// GET /shape/{shapeId}
func ShapePointsHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
	shape_id := r.PathValue("shapeId")

	results := static.GetShapePoints(shape_id)

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

// GET /stopsinfo
func StopsinfoHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)
	results := service.ProcessStopsInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// =======================
// = RAW GTFS ENDPOINTS  =
// =======================

// GET /static/info
func StaticInfoHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)

	min_date, max_date := static.GetCalDatesRange()

	message := "GTFS static data valid from " + min_date + " to " + max_date + "."

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
