package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"

	"github.com/Leocraft1/gtfsparser-with-reader"
)

func GetStopTimes() []model.StopTimesResult {
	var results []model.StopTimesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM stop_times")
	if err != nil {
		fmt.Println("GetStopTimes errore db:", err)
	}

	return results
}

func SaveStopTimes(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	stopTimes := GetStopTimes()

	stopTimesMap := make(map[string]bool)
	for _, val := range stopTimes {
		stopTimesMap[val.Basin+val.Trip_id+val.Stop_id+fmt.Sprint(val.Stop_sequence)] = true
	}

	var new []model.StopTimesResult
	for idx, val := range feedRA.Trips {
		_, ok := stopTimesMap["RA"+idx]
		if !ok {
			for _, val2 := range val.StopTimes {
				newShape := model.ToDomainStopTimes("RA", val.Id, val2.Arrival_time().GetTime(), val2.Departure_time().GetTime(), val2.Stop().Id, val2.Sequence())
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedFC.Trips {
		_, ok := stopTimesMap["FC"+idx]
		if !ok {
			for _, val2 := range val.StopTimes {
				newShape := model.ToDomainStopTimes("FC", val.Id, val2.Arrival_time().GetTime(), val2.Departure_time().GetTime(), val2.Stop().Id, val2.Sequence())
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedRN.Trips {
		_, ok := stopTimesMap["RN"+idx]
		if !ok {
			for _, val2 := range val.StopTimes {
				newShape := model.ToDomainStopTimes("RN", val.Id, val2.Arrival_time().GetTime(), val2.Departure_time().GetTime(), val2.Stop().Id, val2.Sequence())
				new = append(new, newShape)
			}
		}
	}

	//Database insert 
	for _, val := range new {
		_, err := DB_CONTENT.Exec("INSERT INTO stop_times(basin,trip_id,arrival_time,departure_time,stop_id,stop_sequence) VALUES(?, ?, ?, ?, ?, ?)", val.Basin, val.Trip_id, val.Arrival_time, val.Departure_time, val.Stop_id, val.Stop_sequence)
		if err != nil {
			fmt.Println("SaveStopTimes db error:", err)
		}
	}
}