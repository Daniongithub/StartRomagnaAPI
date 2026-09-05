package realtime

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/utils"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func GetStopTimeUpdByTripId(trip_id string) []model.StopTime {
	var results []model.StopTime
	err := repository.DB_RT.Select(&results, "SELECT * FROM stop_time_updates WHERE trip_id = ?", trip_id)
	if err != nil {
		fmt.Println("GetStopTimeUpdByTripId db error:", err)
	}

	return results
}

func GetNextStops(tripId string) model.NextStops {
	var results []model.StopWDel
	err := repository.DB_RT.Select(&results, `
		SELECT stu.delay, st.basin, st.arrival_time, st.departure_time, st.stop_id, st.stop_id, s.stop_code, s.stop_name, s.stop_lat, s.stop_lon FROM stop_time_updates AS stu
		INNER JOIN start_gtfs_static.stop_times AS st
		ON stu.trip_id = st.trip_id AND stu.stop_sequence = st.stop_sequence
		INNER JOIN start_gtfs_static.stops AS s
		ON st.stop_id = s.stop_id
		WHERE stu.trip_id = ? AND s.basin = st.basin
		AND LOCATE('semaforo', stop_name) = 0
		AND LOCATE('fi1', stop_name) = 0
		AND LOCATE('FITTIZIO', stop_name) = 0
		AND LOCATE('Fittizio', stop_name) = 0
		AND LOCATE('FITTIZIA', stop_name) = 0
		AND LOCATE('Fittizia', stop_name) = 0
		ORDER BY st.stop_sequence
	`, tripId)
	if err != nil {
		fmt.Println("GetNextStops db error:", err)
	}

	out := model.NextStops{
		Stops: results,
	}

	return out
}

func GetFirstStop(tripId string) *model.StopWDel {
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
		return nil
	}
	utils.FixStopWDel(results)

	return &results[0]
}

func SaveStopTimeUpd(feeds map[string]*gtfs.FeedMessage) {
	//Database inserts
	for _, val := range feeds {
		for _, val2 := range val.Entity {
			values := make([][]any, 0, len(val2.TripUpdate.StopTimeUpdate))
			for _, val3 := range val2.TripUpdate.StopTimeUpdate {
				if val3.Arrival != nil {
					values = append(values, []any{
						val2.TripUpdate.Trip.TripId,
						val3.StopSequence,
						val3.Arrival.Delay,
					})
				} else {
					values = append(values, []any{
						val2.TripUpdate.Trip.TripId,
						val3.StopSequence,
						val3.Departure.Delay,
					})
				}
			}

			err := repository.BatchInsertTX(repository.TX_TU, "stop_time_updates", []string{"trip_id", "stop_sequence", "delay"}, values)
			if err != nil {
				fmt.Println("SaveStopTimeUpd db error:", err)
			}
		}
	}
}

func DeleteAllStopTimeUpd() {
	_, err := repository.TX_TU.Exec("DELETE FROM stop_time_updates")
	if err != nil {
		fmt.Println("DeleteAllStopTimeUpd db error:", err)
	}
}

func convertDelay(delay *int32) int32 {
	if delay == nil {
		return 0
	}
	out := delay
	return *out
}
