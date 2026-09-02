package realtime

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/utils"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func GetServiceAlerts() []model.ServiceAlertsResult {
	var results []model.ServiceAlertsResult
	err := repository.DB_RT.Select(&results, "SELECT * FROM service_alerts")
	if err != nil {
		fmt.Println("GetServiceAlerts error:", err)
	}

	return results
}

func GetServiceAlertsBasin(basin string) []model.ServiceAlertsResult {
	var results []model.ServiceAlertsResult
	err := repository.DB_RT.Select(&results, "SELECT * FROM service_alerts WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetServiceAlertsBasin error:", err)
	}

	return results
}

func GetDistinctSAByBasin(basin string) []model.ServiceAlertsResult {
	var results []model.ServiceAlertsResult
	err := repository.DB_RT.Select(&results, "SELECT DISTINCT basin, start, end, route_id, route_type, trip_id, direction_id, start_time, start_date FROM service_alerts WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetDistinctSAByBasin error:", err)
	}

	return results
}

func GetFirstStop(tripId string) model.StopWDel {
	var results []model.StopWDel
	err := repository.DB_RT.Select(&results, `
		SELECT stu.delay, st.basin, st.arrival_time, st.departure_time, st.stop_id, st.stop_id, s.stop_code, s.stop_name, s.stop_lat, s.stop_lon FROM stop_time_updates AS stu
		INNER JOIN start_gtfs_static.stop_times AS st
		ON stu.trip_id = st.trip_id
		INNER JOIN start_gtfs_static.stops AS s
		ON st.stop_id = s.stop_id
		WHERE st.basin = s.basin AND stu.trip_id = ? AND st.stop_sequence = (
			SELECT MIN(stu.stop_sequence) FROM stop_time_updates AS stu
			WHERE stu.trip_id = ?
		);
	`, tripId, tripId)
	if err != nil {
		fmt.Println("GetFirstStop error:", err)
	}
	if len(results) == 0 {
		results = append(results,
			model.StopWDel{
				Error: "Trip not available",
			},
		)

		return results[0]
	}
	utils.FixStopWDel(results)

	return results[0]
}

/*
func GetLastStop(tripId string) model.CorseSoppStop {
	var results []model.CorseSoppStop
	err := repository.DB_RT.Select(&results, `
		SELECT DISTINCT n.stop_code, n.stop_name FROM service_alerts AS s
		INNER JOIN start_gtfs_static.stop_times AS t
		ON s.trip_id = t.trip_id
		INNER JOIN start_gtfs_static.stops AS n
		ON t.stop_id = n.stop_id
		WHERE t.basin = n.basin AND s.trip_id = ? AND t.stop_sequence = (
			SELECT MAX(t.stop_sequence) FROM service_alerts AS s
			INNER JOIN start_gtfs_static.stop_times AS t
			ON s.trip_id = t.trip_id
			WHERE s.trip_id = ?
		);
	`, tripId, tripId)
	if err != nil {
		fmt.Println("GetLastStop error:", err)
	}

	return results[0]
}*/

func SaveAlerts(feeds map[string]*gtfs.FeedMessage, dbContent map[string]map[string]bool) {
	rows := make([][]any, 0)

	for idx, val := range feeds {
		for _, val2 := range val.Entity {
			if _, ok := dbContent[idx][*val2.Id]; ok {
				continue
			}

			startS := val2.Alert.ActivePeriod[0].Start
			endS := val2.Alert.ActivePeriod[0].End

			start := secondsToTime(int(*startS))
			end := secondsToTime(int(*endS))

			rows = append(rows, []any{
				idx,
				val2.Id,
				start,
				end,
				val2.Alert.InformedEntity[0].RouteId,
				val2.Alert.InformedEntity[0].RouteType,
				val2.Alert.InformedEntity[0].Trip.TripId,
				val2.Alert.InformedEntity[0].Trip.DirectionId,
				val2.Alert.InformedEntity[0].Trip.StartTime,
				val2.Alert.InformedEntity[0].Trip.StartDate,
				val2.Alert.InformedEntity[0].StopId,
			})
		}
	}

	if len(rows) == 0 {
		return
	}

	columns := []string{
		"basin",
		"id",
		"start",
		"end",
		"route_id",
		"route_type",
		"trip_id",
		"direction_id",
		"start_time",
		"start_date",
		"stop_id",
	}

	err := repository.BatchInsert(repository.DB_RT, "service_alerts", columns, rows)
	if err != nil {
		fmt.Println("SaveAlerts db error:", err)
	}
}

func DeleteAllServiceAlerts() {
	_, err := repository.DB_RT.Exec("DELETE FROM service_alerts")
	if err != nil {
		fmt.Println("DeleteAllServiceAlerts db error:", err)
	}
}

func secondsToTime(seconds int) time.Time {
	base := time.Now()
	midnight := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.UTC)
	return midnight.Add(time.Duration(seconds) * time.Second)
}
