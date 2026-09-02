package realtime

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
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
			fmt.Println("GetTripUpdatesBasin error:", err)
		}
		results[idx].Stop_time_updates = stopTimeUpdate
	}

	return results
}

func GetVehicleByTripId(tripId string) *string {
	var results []string
	err := repository.DB_RT.Select(&results, `SELECT vehicle FROM trip_updates WHERE trip_id = ?`, tripId)
	if err != nil {
		fmt.Println("GetVehicleByTripId error:", err)
	}

	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

func GetVehicles() []string {
	var results []string
	err := repository.DB_RT.Select(&results, `
		SELECT DISTINCT t.vehicle FROM trip_updates AS t
		WHERE t.vehicle != 0 ORDER BY t.vehicle
	`)
	if err != nil {
		fmt.Println("GetVehicles error:", err)
	}

	return results
}

func GetBuses() []model.BusInService {
	var results []model.BusInService
	err := repository.DB_RT.Select(&results, `
		SELECT tu.basin, tu.trip_id, tu.vehicle, tu.timestamp, tu.route_id, t.shape_id, vp.lat, vp.long FROM trip_updates AS tu
		INNER JOIN vehicle_positions AS vp
		ON tu.trip_id = vp.trip_id
		INNER JOIN start_gtfs_static.trips AS t
		ON tu.trip_id = t.trip_id
		ORDER BY tu.basin, tu.route_id
	`)
	if err != nil {
		fmt.Println("GetVehicles error:", err)
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

	err := repository.BatchInsertTX(repository.TX_TU, "trip_updates", []string{"basin", "id", "trip_id", "route_id", "direction_id", "start_time", "start_date", "schedule_relationship", "vehicle", "timestamp"}, values)
	if err != nil {
		fmt.Println("SaveTripUpdates db error:", err)
	}
}

func DeleteAllTripUpdates() {
	_, err := repository.TX_TU.Exec("DELETE FROM trip_updates")
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
