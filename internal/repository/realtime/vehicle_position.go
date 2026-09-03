package realtime

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func GetVehiclePositions() []model.VehiclePosition {
	var results []model.VehiclePosition
	err := repository.DB_RT.Select(&results, `
		SELECT vp.basin, vp.trip_id, vp.vehicle, vp.timestamp, t.route_id, t.shape_id, vp.lat, vp.long FROM vehicle_positions AS vp
		INNER JOIN start_gtfs_static.trips AS t
		ON vp.trip_id = t.trip_id
		ORDER BY vp.basin, t.route_id
	`)
	if err != nil {
		fmt.Println("GetVehicles error:", err)
	}

	return results
}

func GetVehiclePositionsBasin(basin string) []model.VehiclePosition {
	var results []model.VehiclePosition
	err := repository.DB_RT.Select(&results, `
		SELECT vp.basin, vp.trip_id, vp.vehicle, vp.timestamp, t.route_id, t.shape_id, vp.lat, vp.long FROM vehicle_positions AS vp
		INNER JOIN start_gtfs_static.trips AS t
		ON vp.trip_id = t.trip_id
		WHERE vp.basin = ?
		ORDER BY vp.basin, t.route_id
	`, basin)
	if err != nil {
		fmt.Println("GetVehiclePositionsBasin error:", err)
	}

	return results
}

func SaveVehiclePositions(feeds map[string]*gtfs.FeedMessage) {
	values := make([][]any, 0)

	for idx, val := range feeds {
		for _, val2 := range val.Entity {
			values = append(values, []any{
				idx,
				val2.Id,
				val2.Vehicle.Trip.TripId,
				val2.Vehicle.Trip.DirectionId,
				val2.Vehicle.Trip.StartTime,
				val2.Vehicle.Trip.StartDate,
				val2.Vehicle.Vehicle.Label,
				val2.Vehicle.Position.Latitude,
				val2.Vehicle.Position.Longitude,
				time.Unix(int64(*val2.Vehicle.Timestamp), 0),
				val2.Vehicle.OccupancyStatus.String(),
			})
		}
	}

	err := repository.BatchInsertTX(repository.TX_VP, "vehicle_positions", []string{"basin", "id", "trip_id", "direction_id", "start_time", "start_date", "vehicle", "`lat`", "`long`", "timestamp", "occupancy_status"}, values)
	if err != nil {
		fmt.Println("SaveVehiclePositions db error:", err)
	}
}

func DeleteAllVehiclePositions() {
	_, err := repository.TX_VP.Exec("DELETE FROM vehicle_positions")
	if err != nil {
		fmt.Println("DeleteAllVehiclePositions db error:", err)
	}
}
