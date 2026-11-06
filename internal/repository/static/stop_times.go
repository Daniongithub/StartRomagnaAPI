package static

import (
	"fmt"
	"time"

	"startromagnaapi/config"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/utils"
)

func GetStopTimesDetailByTrip(basin, tripId string) []model.StopTimeWName {
	var results []model.StopTimeWName
	err := repository.DB_STATIC.Select(&results, `
		SELECT st.*, s.stop_name FROM stop_times AS st
		INNER JOIN stops AS s
		ON st.stop_id = s.stop_id AND st.basin = s.basin
		WHERE st.basin = ? AND st.trip_id = ?
		ORDER BY st.stop_sequence
	`, basin, tripId)
	if err != nil {
		fmt.Println("GetStopTimesByTrip database error:", err)
	}

	return results
}

func GetTerminusName(basin, tripID string) string {
	var results []string
	err := repository.DB_STATIC.Select(&results, `
		SELECT s.stop_name FROM stop_times st
		INNER JOIN stops s ON st.stop_id = s.stop_id AND st.basin = s.basin
		WHERE st.basin = ? AND st.trip_id = ?
		ORDER BY st.stop_sequence DESC LIMIT 1
	`, basin, tripID)
	if err != nil {
		fmt.Println("GetTerminusName error:", err)
	}
	if len(results) == 0 {
		return ""
	}
	return results[0]
}

func GetArrivals(stopCode string) []model.Arrival {
	var results []model.Arrival
	now := utils.RomeNow()
	endTime := now.Add(time.Duration(config.ARRIVALS_LOAD_INTERVAL) * time.Minute).Format("2006-01-02 15:04:05")
	queryStartTime := now.Add(-time.Duration(config.ARRIVALS_DELAY_BUFFER) * time.Minute).Format("2006-01-02 15:04:05")

	err := repository.DB_STATIC.Select(&results, `
        SELECT 
            st.basin, st.arrival_time, cd.date AS service_date, t.trip_id, t.route_id, t.shape_id, COALESCE(NULLIF(tu.vehicle, 0), vp2.vehicle) AS vehicle, tu.schedule_relationship,
            CASE 
                WHEN vp.trip_id IS NULL THEN TRUE
                WHEN vp.trip_id = tu.trip_id THEN TRUE
                ELSE FALSE
            END AS vehicle_confirmed,
			CASE 
                WHEN vp2.lat IS NOT NULL AND vp2.long IS NOT NULL THEN TRUE
                WHEN vp.trip_id = t.trip_id AND vp.lat IS NOT NULL AND vp.long IS NOT NULL THEN TRUE
                ELSE FALSE
            END AS is_GPS
        FROM stop_times AS st
        INNER JOIN stops AS s
            ON st.stop_id = s.stop_id AND st.basin = s.basin
        INNER JOIN trips AS t
            ON st.trip_id = t.trip_id AND st.basin = t.basin
        INNER JOIN calendar_dates AS cd
            ON t.service_id = cd.service_id AND t.basin = cd.basin
        LEFT JOIN start_gtfs_rt.trip_updates AS tu
            ON st.trip_id = tu.trip_id AND st.basin = tu.basin AND tu.start_date = cd.date
		LEFT JOIN start_gtfs_rt.vehicle_positions AS vp
			ON tu.vehicle = vp.vehicle AND st.basin = vp.basin
		LEFT JOIN start_gtfs_rt.vehicle_positions AS vp2
			ON st.trip_id = vp2.trip_id AND st.basin = vp2.basin AND vp2.start_date = cd.date
        WHERE s.stop_code = ? AND cd.exception_type = 1
 AND cd.date BETWEEN DATE_SUB(DATE(?), INTERVAL 6 DAY) AND DATE(?)
 AND TIMESTAMP(cd.date, st.arrival_time) >= ? AND TIMESTAMP(cd.date, st.arrival_time) <= ?
        ORDER BY cd.date, st.arrival_time;
    `, stopCode, queryStartTime, endTime, queryStartTime, endTime)
	if err != nil {
		fmt.Println("GetLastStop error:", err)
	}

	return results
}
