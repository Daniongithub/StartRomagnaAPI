package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"time"

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
		fmt.Println("GetStopTimesBasin errore db:", err)
	}

	return results
}

func SaveStopTimes(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	stopTimes := GetStopTimes()

	stopTimesMap := make(map[string]bool)
	for _, val := range stopTimes {
		stopTimesMap[val.Basin+val.Trip_id] = true
	}

	now := time.Now()

	var new []model.StopTimesResult
	var old []model.StopTimesResult
	feedKeys := make(map[string]bool)
	for idx, val := range feedRA.Trips {
		_, ok := stopTimesMap["RA"+idx]
		feedKeys["RA"+idx] = true
		if !ok {
			for _, val2 := range val.StopTimes {
				var gdate gtfs.Date
				gdate.SetYear(uint16(now.Year()))
				gdate.SetMonth(uint8(now.Month()))
				gdate.SetDay(uint8(now.Day()))
				var fakeAg gtfs.Agency
				fakeAg.Timezone = *time.UTC
				newShape := model.ToDomainStopTimes("RA", val.Id, val2.Arrival_time().GetLocationTime(gdate, &fakeAg), val2.Departure_time().GetLocationTime(gdate, &fakeAg), val2.Stop().Id, val2.Sequence())
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedFC.Trips {
		_, ok := stopTimesMap["FC"+idx]
		feedKeys["FC"+idx] = true
		if !ok {
			for _, val2 := range val.StopTimes {
				var gdate gtfs.Date
				gdate.SetYear(uint16(now.Year()))
				gdate.SetMonth(uint8(now.Month()))
				gdate.SetDay(uint8(now.Day()))
				var fakeAg gtfs.Agency
				fakeAg.Timezone = *time.UTC
				newShape := model.ToDomainStopTimes("FC", val.Id, val2.Arrival_time().GetLocationTime(gdate, &fakeAg), val2.Departure_time().GetLocationTime(gdate, &fakeAg), val2.Stop().Id, val2.Sequence())
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedRN.Trips {
		_, ok := stopTimesMap["RN"+idx]
		feedKeys["RN"+idx] = true
		if !ok {
			for _, val2 := range val.StopTimes {
				var gdate gtfs.Date
				gdate.SetYear(uint16(now.Year()))
				gdate.SetMonth(uint8(now.Month()))
				gdate.SetDay(uint8(now.Day()))
				var fakeAg gtfs.Agency
				fakeAg.Timezone = *time.UTC
				newShape := model.ToDomainStopTimes("RN", val.Id, val2.Arrival_time().GetLocationTime(gdate, &fakeAg), val2.Departure_time().GetLocationTime(gdate, &fakeAg), val2.Stop().Id, val2.Sequence())
				new = append(new, newShape)
			}
		}
	}

	for _, val := range stopTimes {
		_, ok := feedKeys[val.Basin+val.Trip_id]
		if !ok {
			var oldStopTime model.StopTimesResult
			oldStopTime.Basin = val.Basin
			oldStopTime.Trip_id = val.Trip_id
			old = append(old, oldStopTime)
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

	//Database delete
	for _, val := range old {
		_, err = repository.DB_STATIC.Exec("DELETE FROM stop_times WHERE basin = ? AND trip_id = ?", val.Basin, val.Trip_id)
		if err != nil {
			fmt.Println("SaveStopTimes db error:", err)
		}
	}
}
