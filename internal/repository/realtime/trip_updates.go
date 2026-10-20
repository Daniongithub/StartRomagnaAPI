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

func GetVehicles() []model.VehicleReport {
	var results []model.VehicleReport
	err := repository.DB_RT.Select(&results, `SELECT DISTINCT vehicle, basin FROM trip_updates WHERE vehicle != 0 ORDER BY vehicle`)
	if err != nil {
		fmt.Println("GetVehicles error:", err)
	}

	return results
}

func GetBuses() []model.BusInService {
	var results []model.BusInService
	err := repository.DB_RT.Select(&results, `
		WITH tu_filtered AS (
			SELECT
				tu.basin, tu.trip_id, tu.vehicle, tu.timestamp,
				TIMESTAMP(tu.start_date, tu.start_time) AS start_datetime
			FROM trip_updates AS tu
			WHERE tu.start_time < NOW()
			AND (tu.schedule_relationship <> 'CANCELED' OR tu.schedule_relationship IS NULL)
			-- AND tu.basin = ?  -- se filtri per basin
		),
		tu_trip AS (
			SELECT basin, vehicle, trip_id,
				MAX(start_datetime) AS start_datetime,
				MAX(timestamp) AS timestamp
			FROM tu_filtered
			GROUP BY basin, vehicle, trip_id
		),
		tu_best AS (
			SELECT basin, vehicle, trip_id, start_datetime, timestamp
			FROM (
				SELECT basin, vehicle, trip_id, start_datetime, timestamp,
					ROW_NUMBER() OVER (PARTITION BY vehicle ORDER BY start_datetime DESC) AS rn
				FROM tu_trip
			) x
			WHERE rn = 1
		),
		vp_trip AS (
			SELECT vp.basin, vp.vehicle, vp.trip_id, vp.lat, vp.long AS longitude, vp.timestamp
			FROM vehicle_positions AS vp
			-- WHERE vp.basin = ?  -- se filtri per basin
		),
		resolved_with_vp AS (
			SELECT
				v.basin,
				v.vehicle,
				CASE
					WHEN b.trip_id IS NULL OR b.trip_id = v.trip_id THEN v.trip_id
					WHEN vs.start_datetime IS NOT NULL AND vs.start_datetime >= b.start_datetime THEN v.trip_id
					ELSE b.trip_id
				END AS trip_id,
				CASE
					WHEN b.trip_id IS NULL OR b.trip_id = v.trip_id THEN v.lat
					WHEN vs.start_datetime IS NOT NULL AND vs.start_datetime >= b.start_datetime THEN v.lat
					ELSE NULL
				END AS lat,
				CASE
					WHEN b.trip_id IS NULL OR b.trip_id = v.trip_id THEN v.longitude
					WHEN vs.start_datetime IS NOT NULL AND vs.start_datetime >= b.start_datetime THEN v.longitude
					ELSE NULL
				END AS longitude,
				v.timestamp
			FROM vp_trip AS v
			LEFT JOIN tu_best AS b ON b.vehicle = v.vehicle
			LEFT JOIN tu_trip AS vs ON vs.vehicle = v.vehicle AND vs.trip_id = v.trip_id
		),
		resolved_tu_only AS (
			SELECT b.basin, b.vehicle, b.trip_id, NULL AS lat, NULL AS longitude, b.timestamp
			FROM tu_best AS b
			WHERE NOT EXISTS (
				SELECT 1 FROM vehicle_positions AS vp WHERE vp.vehicle = b.vehicle
			)
		),
		resolved AS (
			SELECT * FROM resolved_with_vp
			UNION ALL
			SELECT * FROM resolved_tu_only
		)
		SELECT
			r.basin, r.trip_id, r.vehicle, r.timestamp,
			t.route_id, t.shape_id,
			r.lat, r.longitude
		FROM resolved AS r
		INNER JOIN start_gtfs_static.trips AS t
		ON r.trip_id = t.trip_id
	`)
	if err != nil {
		fmt.Println("GetBuses error:", err)
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

func GetBus(id string) []model.BusInService {
	var results []model.BusInService
	err := repository.DB_RT.Select(&results, `
		SELECT vp.basin, vp.trip_id, vp.vehicle, vp.timestamp, t.route_id, t.shape_id, vp.lat, vp.long
		FROM vehicle_positions AS vp
		INNER JOIN start_gtfs_static.trips AS t
		ON vp.trip_id = t.trip_id
		WHERE vehicle = ?
		UNION ALL

		SELECT ranked.basin, ranked.trip_id, ranked.vehicle, ranked.timestamp, t.route_id, t.shape_id, NULL AS lat, NULL AS 'long'
		FROM (
			SELECT tu.basin, tu.trip_id, tu.vehicle, tu.timestamp, tu.start_time, tu.schedule_relationship,
				ROW_NUMBER() OVER (
					PARTITION BY tu.vehicle
					ORDER BY tu.timestamp ASC
				) AS rn
			FROM trip_updates AS tu
			WHERE tu.start_time < NOW()
			AND (tu.schedule_relationship <> 'CANCELED' OR tu.schedule_relationship IS NULL)
		) AS ranked
		INNER JOIN start_gtfs_static.trips AS t
			ON ranked.trip_id = t.trip_id
		WHERE ranked.rn = 1
		AND NOT EXISTS (
				SELECT 1
				FROM vehicle_positions AS vp2
				WHERE vp2.trip_id = ranked.trip_id
			)
		AND vehicle = ?
		ORDER BY basin, route_id
	`, id, id)
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
