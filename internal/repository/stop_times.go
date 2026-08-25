package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"

	"github.com/Leocraft1/gtfsparser-with-reader"
	"github.com/Leocraft1/gtfsparser-with-reader/gtfs"
	"github.com/Masterminds/squirrel"
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
		stopTimesMap[val.Basin+val.Trip_id] = true
	}

	var new []model.StopTimesResult
	for idx, val := range feedRA.Trips {
		_, ok := stopTimesMap["RA"+idx]
		if !ok {
			for _, val2 := range val.StopTimes {
				var gdate gtfs.Date
				gdate.SetYear(2010)
				gdate.SetMonth(9)
				gdate.SetDay(11)
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

	values := make([][]interface{}, 0, len(stopTimes))

	for _, val := range stopTimes {
		values = append(values, []interface{}{
			val.Basin,
			val.Trip_id,
			val.Arrival_time,
			val.Departure_time,
			val.Stop_id,
			val.Stop_sequence,
		})
	}

	//Database insert
	const batchSize = 5000

	for start := 0; start < len(new); start += batchSize {
		end := start + batchSize
		if end > len(new) {
			end = len(new)
		}

		q := squirrel.Insert("stop_times").
			Columns(
				"basin",
				"trip_id",
				"arrival_time",
				"departure_time",
				"stop_id",
				"stop_sequence",
			)

		for _, val := range new[start:end] {
			q = q.Values(
				val.Basin,
				val.Trip_id,
				val.Arrival_time,
				val.Departure_time,
				val.Stop_id,
				val.Stop_sequence,
			)
		}

		sql, args, _ := q.ToSql()

		if _, err := DB_CONTENT.Exec(sql, args...); err != nil {
			fmt.Println("SaveStopTimes db error:", err)
		}
	}
}
