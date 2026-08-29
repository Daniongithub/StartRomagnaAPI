package realtime

import (
	"StartRomagnaAPI/internal/model"
	"StartRomagnaAPI/internal/repository"
	"fmt"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func GetTripUpdatesBasin(basin string) []model.TripUpdatesResult {
	var results []model.TripUpdatesResult
	err := repository.DB_RT.Select(&results, "SELECT * FROM trip_updates WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetTripUpdatesBasin error:", err)
	}
	for idx, val := range results {
		var stopTimeUpdate []model.StopTime
		err = repository.DB_RT.Select(&stopTimeUpdate, "SELECT * FROM stop_time_update WHERE trip_id = ?", val.Trip_id)
		if err != nil {
			fmt.Println("GetTripUpdates error:", err)
		}
		results[idx].Stop_time_updates = stopTimeUpdate
	}

	return results
}

func SaveTripUpdates(feeds map[string]*gtfs.FeedMessage) {
	//Database inserts
	for idx, val := range feeds {
		for _, val2 := range val.Entity {
			schedRel := getSchedRel(val2)
			_, err := repository.DB_RT.Exec("INSERT INTO trip_updates VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", idx, val2.Id, val2.TripUpdate.Trip.TripId, val2.TripUpdate.Trip.RouteId, val2.TripUpdate.Trip.DirectionId, val2.TripUpdate.Trip.StartTime, val2.TripUpdate.Trip.StartDate, schedRel, val2.TripUpdate.Vehicle.Label, time.Unix(int64(*val2.TripUpdate.Timestamp), 0))
			if err != nil {
				fmt.Println("SaveTripUpdates db error:", err)
			}
		}
	}
}

func DeleteAllTripUpdates() {
	_, err := repository.DB_RT.Exec("DELETE FROM trip_updates")
	if err != nil {
		fmt.Println("DeleteAllTripUpdates db error:", err)
	}
}

func getSchedRel(val2 *gtfs.FeedEntity) *string{
	if val2.TripUpdate.Trip.ScheduleRelationship == nil {
		return nil;
	}
	out := val2.TripUpdate.Trip.ScheduleRelationship.String()
	return &out;
}