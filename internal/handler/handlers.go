package handler

import (
	"StartRomagnaAPI/config"
	//"StartRomagnaAPI/internal/auth"
	"StartRomagnaAPI/internal/gtfs"
	"StartRomagnaAPI/internal/repository"
	"encoding/json"
	"fmt"
	//"io"
	"net/http"
	//"sort"

	//"google.golang.org/protobuf/proto"
)

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	message := "API is healthy and running on port " + config.PORT

	w.Write([]byte(message))
}

/**
* Very experimental handler: using for testing purpose only
**/
func TripsHandler(w http.ResponseWriter, r *http.Request) {
	results := repository.GetTrips()

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


func PasteBin(w http.ResponseWriter, r *http.Request) {
	//urlRA := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Ravenna.zip"
	urlFC := config.START_GTFS_ROOT + "/AVM/GTFSStatic_FOCE.zip"
	urlRN := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Rimini.zip"

	/*feedRA, err := gtfs.GetStaticFeed(urlRA)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}*/
	feedFC, err := gtfs.GetStaticFeed(urlFC)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}
	feedRN, err := gtfs.GetStaticFeed(urlRN)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}
	out := []string{feedFC.Trips["F_R956592"].Route.Id, feedRN.Trips["RN4093119"].Route.Id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
	//fmt.Println(feedRA.Trips["RA2889361"].Route.Id)
	//fmt.Println(feedFC.Trips["F_R956592"].Route.Id)
	//fmt.Println(feedRN.Trips["RN4093119"].Route.Id)
}
