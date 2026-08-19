package handler

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/auth"
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	gtfs "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	message := "API is healthy and running on port " + config.PORT

	w.Write([]byte(message))
}

/**
* Very experimental handler: using for testing purpose only
**/
func GTFSHandler(w http.ResponseWriter, r *http.Request) {
	url := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Ravenna.zip"

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

	w.WriteHeader(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	reader := bytes.NewReader(body)

	zipReader, err := zip.NewReader(reader, int64(len(body)))
	if err != nil {
		return
	}

	for _, file := range zipReader.File {
		if file.Name != "trips.txt" {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return
		}
		defer rc.Close()

		csvReader := csv.NewReader(rc)

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}

			fmt.Println(record)
		}
	}
}

func VehiclesHandler(w http.ResponseWriter, r *http.Request) {
	url := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-vehicle-positions-ra.pb"

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

	type Vehicle struct {
		ID        string  `json:"id"`
		Label     string  `json:"label"`
		TripID    string  `json:"trip_id"`
		StartTime string  `json:"start_time"`
		StartDate string  `json:"start_date"`
		Latitude  float32 `json:"latitude"`
		Longitude float32 `json:"longitude"`
	}

	vehicles := make([]Vehicle, 0)

	for _, entity := range feed.Entity {
		v := entity.GetVehicle()
		if v == nil {
			continue
		}

		vehicles = append(vehicles, Vehicle{
			ID:        v.GetVehicle().GetId(),
			Label:     v.GetVehicle().GetLabel(),
			TripID:    v.GetTrip().GetTripId(),
			StartTime: v.GetTrip().GetStartTime(),
			StartDate: v.GetTrip().GetStartDate(),
			Latitude:  v.GetPosition().GetLatitude(),
			Longitude: v.GetPosition().GetLongitude(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicles)
}
