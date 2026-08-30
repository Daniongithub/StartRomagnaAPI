package realtime

import (
	"StartRomagnaAPI/internal/model"
	"StartRomagnaAPI/internal/repository"
	"fmt"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/jmoiron/sqlx"
)

var TX *sqlx.Tx

func GetStopTimeUpdByTripId(trip_id string) []model.StopTime {
	var results []model.StopTime
	err := repository.DB_RT.Select(&results, "SELECT * FROM stop_time_updates WHERE trip_id = ?", trip_id)
	if err != nil {
		fmt.Println("GetStopTimeUpdByTripId db error:", err)
	}

	return results
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

			err := repository.BatchInsertTX(TX, "stop_time_updates", []string{"trip_id", "stop_sequence", "delay"}, values)
			if err != nil {
				fmt.Println("SaveStopTimeUpd db error:", err)
			}
		}
	}
}

func DeleteAllStopTimeUpd() {
	_, err := TX.Exec("DELETE FROM stop_time_updates")
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

func StartTransaction() {
	var err error
	TX, err = repository.DB_RT.Beginx()
	if err != nil {
		fmt.Println("StartTransaction db error:", err)
	}
}

func CommitTransaction() {
	err := TX.Commit()
	if err != nil {
		fmt.Println("CommitTransaction db error:", err)
	}
}
