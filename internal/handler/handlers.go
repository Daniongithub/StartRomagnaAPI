package handler

import (
	"encoding/json"
	"net/http"
	"startromagnaapi/config"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/service"
	"time"

	"github.com/mmcdole/gofeed"
)

const feedURL = "https://www.startromagna.it/infobus/feed/"

// GET /
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)

	message := "API is healthy and running on port " + config.PORT

	w.Write([]byte(message))
}

// GET /rss/feed
func RSSFeedHandler(w http.ResponseWriter, r *http.Request) {
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

	AddCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)
}

// GET /timetable/{routeid}
/*
func TimetableHandler(w http.ResponseWriter, r *http.Request) {
	routeId := r.PathValue("routeid")

	results := service.ProcessTimetable(routeId)

	AddCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}*/

// GET /corsesopp
func CorsesoppHandler(w http.ResponseWriter, r *http.Request) {
	results := service.ProcessCorsesopp("RA")

	AddCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// ------------------------
// - RAW GTFS ENDPOINTS
// ------------------------

// GET /static/info
func StaticInfoHandler(w http.ResponseWriter, r *http.Request) {
	AddCORS(w, r)

	min_date, max_date := static.GetCalDatesRange()

	message := "GTFS static data valid from " + min_date + " to " + max_date + "."

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

	AddCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /static/calendar_dates/{basin}
func CalDatesBasinHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetCalDatesBasin(basin)

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

	results := static.GetRoutesBasin(basin)

	AddCORS(w, r)
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

	AddCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /static/stop_times/{basin}
func StopTimesBasinHandler(w http.ResponseWriter, r *http.Request) {
	basin := r.PathValue("basin")

	if basin != "RA" && basin != "FC" && basin != "RN" {
		http.Error(w, "Invalid basin", http.StatusBadRequest)
		return
	}

	results := static.GetStopTimesBasin(basin)

	AddCORS(w, r)
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

	AddCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

/*
func RTHandler(w http.ResponseWriter, r *http.Request) {
	url := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-trip-updates-ra.pb"

	req, err := auth.BasicAuth("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "GTFS-RT server returned "+resp.Status, http.StatusBadGateway)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	var feed gtfs.FeedMessage

	if err := proto.Unmarshal(body, &feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Trip struct {
		ID                   string `json:"id"`
		VehicleID            string `json:"vehicle_id"`
		VehicleLabel         string `json:"vehicle_label"`
		TripID               string `json:"trip_id"`
		RouteID              string `json:"route_id"`
		StartTime            string `json:"start_time"`
		StartDate            string `json:"start_date"`
		ScheduleRelationship string `json:"schedule_relationship"`
	}

	trips := make([]Trip, 0)

	for _, entity := range feed.Entity {
		tu := entity.GetTripUpdate()
		if tu == nil {
			continue
		}

		trip := tu.GetTrip()
		vehicle := tu.GetVehicle()

		trips = append(trips, Trip{
			ID:                   entity.GetId(),
			VehicleID:            vehicle.GetId(),
			VehicleLabel:         vehicle.GetLabel(),
			TripID:               trip.GetTripId(),
			RouteID:              trip.GetRouteId(),
			StartTime:            trip.GetStartTime(),
			StartDate:            trip.GetStartDate(),
			ScheduleRelationship: trip.GetScheduleRelationship().String(),
		})
	}

	sort.Slice(trips, func(i, j int) bool {
		return trips[i].VehicleLabel < trips[j].VehicleLabel
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}*/
