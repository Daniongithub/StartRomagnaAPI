package realtime

import (
	"fmt"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/utils"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/jmoiron/sqlx"
)

func GetNextStops(tripId string) model.NextStops {
	var results []model.StopWDel
	err := repository.DB_RT.Select(&results, `
		SELECT stu.delay, st.basin, st.arrival_time, st.departure_time, st.stop_id, s.stop_code, s.stop_name, s.stop_lat, s.stop_lon FROM stop_time_updates AS stu
		INNER JOIN start_gtfs_static.stop_times AS st
		ON stu.trip_id = st.trip_id AND stu.stop_sequence = st.stop_sequence
		INNER JOIN start_gtfs_static.stops AS s
		ON st.stop_id = s.stop_id
		WHERE stu.trip_id = ? AND s.basin = st.basin
		AND s.is_dummy = 0
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
        SELECT stu.delay, st.basin, st.arrival_time, st.departure_time, st.stop_id, s.stop_code, s.stop_name, s.stop_lat, s.stop_lon
        FROM stop_time_updates AS stu
        INNER JOIN start_gtfs_static.stop_times AS st
            ON stu.trip_id = st.trip_id AND stu.stop_sequence = st.stop_sequence
        INNER JOIN start_gtfs_static.stops AS s
            ON st.stop_id = s.stop_id AND st.basin = s.basin
        WHERE stu.trip_id = ?
        AND s.is_dummy = 0
        ORDER BY st.stop_sequence ASC
        LIMIT 1
    `, tripId)
	if err != nil {
		fmt.Println("GetFirstStop error:", err)
	}
	if len(results) == 0 {
		return nil
	}
	utils.FixStopWDel(results)

	return &results[0]
}

func SaveStopTimeUpd(tx *sqlx.Tx, feeds map[string]*gtfs.FeedMessage) error {
	values := make([][]any, 0)
	for _, feed := range feeds {
		for _, entity := range feed.GetEntity() {
			update := entity.GetTripUpdate()
			if entity.GetIsDeleted() || update == nil {
				continue
			}
			for _, stop := range update.GetStopTimeUpdate() {
				// Stop ID only and absolute time only need resolution against the static feed.
				// Do not invent a zero delay or dereference absent departure events.
				if stop == nil || stop.StopSequence == nil {
					continue
				}
				var delay *int32
				if stop.GetArrival() != nil {
					delay = stop.GetArrival().Delay
				}
				if delay == nil && stop.GetDeparture() != nil {
					delay = stop.GetDeparture().Delay
				}
				if delay == nil {
					continue
				}
				values = append(values, []any{update.GetTrip().GetTripId(), stop.GetStopSequence(), *delay})
			}
		}
	}
	return repository.BatchInsert(tx, "stop_time_updates", []string{"trip_id", "stop_sequence", "delay"}, values)
}
