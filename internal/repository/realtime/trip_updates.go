package realtime

import (
	"fmt"
	"time"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/utils"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/jmoiron/sqlx"
)

func GetVehicles() []model.VehicleReport {
	var results []model.VehicleReport
	err := repository.DB_RT.Select(&results, `SELECT DISTINCT vehicle, basin FROM trip_updates WHERE vehicle != 0 ORDER BY vehicle`)
	if err != nil {
		fmt.Println("GetVehicles error:", err)
	}

	return results
}

func GetBuses() []model.BusInService        { return getBuses(nil) }
func GetBus(id string) []model.BusInService { return getBuses(&id) }

func getBuses(id *string) []model.BusInService {
	var results []model.BusInService
	query := `
		WITH tu_filtered AS (
			SELECT
				tu.basin, tu.trip_id, tu.vehicle, tu.timestamp,
				TIMESTAMP(tu.start_date, tu.start_time) AS start_datetime
			FROM trip_updates AS tu
			WHERE TIMESTAMP(tu.start_date, tu.start_time) <= ?
 AND tu.vehicle <> '0' AND tu.vehicle <> ''
			AND (tu.schedule_relationship <> 'CANCELED' OR tu.schedule_relationship IS NULL)
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
			r.lat, r.longitude AS "long"
		FROM resolved AS r
		INNER JOIN start_gtfs_static.trips AS t
		ON r.trip_id = t.trip_id AND r.basin = t.basin
 `
	args := []any{utils.RomeNow().Format("2006-01-02 15:04:05")}
	if id != nil {
		query += " WHERE r.vehicle = ?"
		args = append(args, *id)
	}
	query += " ORDER BY r.basin, t.route_id, r.vehicle"
	err := repository.DB_RT.Select(&results, query, args...)
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

func SaveTripUpdates(tx *sqlx.Tx, feeds map[string]*gtfs.FeedMessage) error {
	values := make([][]any, 0)
	for basin, feed := range feeds {
		for _, entity := range feed.GetEntity() {
			update := entity.GetTripUpdate()
			if entity.GetIsDeleted() || update == nil {
				continue
			}
			trip := update.GetTrip()
			if trip.GetTripId() == "" {
				return fmt.Errorf("Trip update %s: missing trip id", entity.GetId())
			}
			stamp := update.GetTimestamp()
			if stamp == 0 {
				stamp = feed.GetHeader().GetTimestamp()
			}
			var schedRel *string
			if trip.ScheduleRelationship != nil {
				value := trip.ScheduleRelationship.String()
				schedRel = &value
			}
			var vehicleLabel *string
			if vehicle := update.GetVehicle(); vehicle != nil {
				vehicleLabel = vehicle.Label
			}
			values = append(values, []any{basin, entity.GetId(), trip.GetTripId(), trip.RouteId, trip.DirectionId,
				trip.StartTime, trip.StartDate, schedRel, vehicleLabel, time.Unix(int64(stamp), 0).UTC()})
		}
	}
	return repository.BatchInsert(tx, "trip_updates", []string{"basin", "id", "trip_id", "route_id", "direction_id", "start_time", "start_date", "schedule_relationship", "vehicle", "timestamp"}, values)
}
