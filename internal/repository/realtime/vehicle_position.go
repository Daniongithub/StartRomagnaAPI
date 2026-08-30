package realtime

import (
	"StartRomagnaAPI/internal/repository"
	"fmt"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

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

	err := repository.BatchInsertTX(TX, "vehicle_positions", []string{"basin", "id", "trip_id", "direction_id", "start_time", "start_date", "vehicle", "`lat`", "`long`", "timestamp", "occupancy_status"}, values)
	if err != nil {
		fmt.Println("SaveVehiclePositions db error:", err)
	}
}

func DeleteAllVehiclePositions() {
	_, err := TX.Exec("DELETE FROM vehicle_positions")
	if err != nil {
		fmt.Println("DeleteAllVehiclePositions db error:", err)
	}
}