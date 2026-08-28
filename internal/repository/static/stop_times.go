package static

import (
	"StartRomagnaAPI/internal/model"
	"StartRomagnaAPI/internal/repository"
	"fmt"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
	"github.com/Leocraft1/gtfsparser-with-reader/gtfs"
)

func GetStopTimes() []model.StopTimesResult {
	var results []model.StopTimesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM stop_times")
	if err != nil {
		fmt.Println("GetStopTimes errore db:", err)
	}

	return results
}

func GetStopTimesBasin(basin string) []model.StopTimesResult {
	var results []model.StopTimesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM stop_times WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetStopTimes errore db:", err)
	}

	return results
}

func SaveStopTimes(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	stopTimes := GetStopTimes()

	stopTimesMap := make(map[string]bool)
	for _, val := range stopTimes {
		stopTimesMap[val.Basin+val.Trip_id] = true
	}

	var new []model.StopTimesResult
	for idx, val := range feedRA.Trips {
		_, ok := stopTimesMap["RA"+idx]
		if !ok {
			for _, val2 := range val.StopTimes {
				var gdate gtfs.Date
				gdate.SetYear(2000)
				gdate.SetMonth(1)
				gdate.SetDay(1)
				newShape := model.ToDomainStopTimes("RA", val.Id, val2.Arrival_time().GetLocationTime(gdate, feedRA.Agencies["START RA"]), val2.Departure_time().GetLocationTime(gdate, feedRA.Agencies["START RA"]), val2.Stop().Id, val2.Sequence())
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
	values := make([][]any, 0, len(new))

	for _, val := range new {
		values = append(values, []any{
			val.Basin,
			val.Trip_id,
			val.Arrival_time,
			val.Departure_time,
			val.Stop_id,
			val.Stop_sequence,
		})
	}

	err := repository.BatchInsert(repository.DB_STATIC, "stop_times", []string{"basin", "trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence"}, values)
	if err != nil {
		fmt.Println("SaveStopTimes db error:", err)
	}

}
