package realtime

import (
	"fmt"
	"time"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/jmoiron/sqlx"
)

func GetVehiclePositions() []model.BusInService {
	var results []model.BusInService
	err := repository.DB_RT.Select(&results, `
		SELECT vp.basin, vp.trip_id, vp.vehicle, vp.timestamp, t.route_id, t.shape_id, vp.lat, vp.long FROM vehicle_positions AS vp
		INNER JOIN start_gtfs_static.trips AS t
		ON vp.trip_id = t.trip_id AND vp.basin = t.basin
		ORDER BY vp.basin, t.route_id
	`)
	if err != nil {
		fmt.Println("GetVehicles error:", err)
	}

	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		fmt.Println("LoadLocation error:", err)
		return results
	}

	for i := range results {
		results[i].LastUpdate = results[i].LastUpdate.In(loc)
	}

	return results
}

func GetVehiclePositionsBasin(basin string) []model.BusInService {
	var results []model.BusInService
	err := repository.DB_RT.Select(&results, `
		SELECT vp.basin, vp.trip_id, vp.vehicle, vp.timestamp, t.route_id, t.shape_id, vp.lat, vp.long FROM vehicle_positions AS vp
		INNER JOIN start_gtfs_static.trips AS t
		ON vp.trip_id = t.trip_id AND vp.basin = t.basin
		WHERE vp.basin = ?
		ORDER BY vp.basin, t.route_id
	`, basin)
	if err != nil {
		fmt.Println("GetVehiclePositionsBasin error:", err)
	}

	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		fmt.Println("LoadLocation error:", err)
		return results
	}

	for i := range results {
		results[i].LastUpdate = results[i].LastUpdate.In(loc)
	}

	return results
}

func GetVehicleLocationID(id string) *model.BusInService {
	var results []model.BusInService
	err := repository.DB_RT.Select(&results, `
		SELECT vp.basin, vp.trip_id, vp.vehicle, vp.timestamp, t.route_id, t.shape_id, vp.lat, vp.long FROM vehicle_positions AS vp
		INNER JOIN start_gtfs_static.trips AS t
		ON vp.trip_id = t.trip_id AND vp.basin = t.basin
		WHERE vp.vehicle = ?
	`, id)
	if err != nil {
		fmt.Println("GetVehiclePositionsBasin error:", err)
	}

	if len(results) == 0 {
		return nil
	}

	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		fmt.Println("LoadLocation error:", err)
		return &results[0]
	}

	results[0].LastUpdate = results[0].LastUpdate.In(loc)

	return &results[0]
}

func SaveVehiclePositions(tx *sqlx.Tx, feeds map[string]*gtfs.FeedMessage) error {
	values := make([][]any, 0)
	for basin, feed := range feeds {
		for _, entity := range feed.GetEntity() {
			vehicle := entity.GetVehicle()
			if entity.GetIsDeleted() || vehicle == nil {
				continue
			}
			trip := vehicle.GetTrip()
			// A vehicle can be validly unassigned to a trip. These endpoints expose trips only.
			if trip.GetTripId() == "" || vehicle.GetPosition() == nil {
				continue
			}
			stamp := vehicle.GetTimestamp()
			if stamp == 0 {
				stamp = feed.GetHeader().GetTimestamp()
			}
			var occupancy *string
			if vehicle.OccupancyStatus != nil {
				value := vehicle.OccupancyStatus.String()
				occupancy = &value
			}
			var vehicleLabel *string
			if descriptor := vehicle.GetVehicle(); descriptor != nil {
				vehicleLabel = descriptor.Label
			}
			values = append(values, []any{basin, entity.GetId(), trip.GetTripId(), trip.DirectionId,
				trip.StartTime, trip.StartDate, vehicleLabel, vehicle.GetPosition().GetLatitude(),
				vehicle.GetPosition().GetLongitude(), time.Unix(int64(stamp), 0).UTC(), occupancy})
		}
	}
	return repository.BatchInsert(tx, "vehicle_positions", []string{"basin", "id", "trip_id", "direction_id", "start_time", "start_date", "vehicle", "`lat`", "`long`", "timestamp", "occupancy_status"}, values)
}
