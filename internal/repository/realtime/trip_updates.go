package realtime

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
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
	values := make([][]any, 0)

	for idx, feed := range feeds {
		for _, entity := range feed.Entity {
			schedRel := getSchedRel(entity)

			values = append(values, []any{
				idx,
				entity.Id,
				entity.TripUpdate.Trip.TripId,
				entity.TripUpdate.Trip.RouteId,
				entity.TripUpdate.Trip.DirectionId,
				entity.TripUpdate.Trip.StartTime,
				entity.TripUpdate.Trip.StartDate,
				schedRel,
				entity.TripUpdate.Vehicle.Label,
				time.Unix(int64(*entity.TripUpdate.Timestamp), 0),
			})
		}
	}

	err := repository.BatchInsertTX(TX, "trip_updates", []string{"basin", "id", "trip_id", "route_id", "direction_id", "start_time", "start_date", "schedule_relationship", "vehicle", "timestamp"}, values)
	if err != nil {
		fmt.Println("SaveTripUpdates db error:", err)
	}
}

func DeleteAllTripUpdates() {
	_, err := TX.Exec("DELETE FROM trip_updates")
	if err != nil {
		fmt.Println("DeleteAllTripUpdates db error:", err)
	}
}

func getSchedRel(val2 *gtfs.FeedEntity) *string {
	if val2.TripUpdate.Trip.ScheduleRelationship == nil {
		return nil
	}
	out := val2.TripUpdate.Trip.ScheduleRelationship.String()
	return &out
}
