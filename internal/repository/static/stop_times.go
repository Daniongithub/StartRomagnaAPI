package static

import (
	"fmt"
	"startromagnaapi/config"
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

func GetTerminusName(tripId string) string {
	var results []string
	err := repository.DB_STATIC.Select(&results, `
		SELECT s.stop_name FROM stop_times AS st
		INNER JOIN start_gtfs_static.stops AS s
		ON st.stop_id = s.stop_id
		WHERE st.basin = s.basin AND st.trip_id = ? AND st.stop_sequence = (
			SELECT MAX(stop_sequence) FROM stop_times
			WHERE trip_id = ?
		)
	`, tripId, tripId)
	if err != nil {
		fmt.Println("GetLastStop error:", err)
	}

	return results[0]
}

func GetArrivals(stopCode string) []model.Arrival {
	var results []model.Arrival
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
	cdDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Duration(config.ARRIVALS_LOAD_INTERVAL) * time.Minute)
	err := repository.DB_STATIC.Select(&results, `
		SELECT st.basin, st.arrival_time, st.trip_id, t.route_id, t.shape_id, tu.vehicle FROM stop_times AS st
		INNER JOIN stops AS s
		ON st.stop_id = s.stop_id AND st.basin = s.basin
		INNER JOIN trips AS t
		ON st.trip_id = t.trip_id AND st.basin = t.basin
		INNER JOIN calendar_dates AS cd
		ON t.service_id = cd.service_id AND t.basin = cd.basin
		LEFT JOIN start_gtfs_rt.trip_updates AS tu
		ON st.trip_id = tu.trip_id AND st.basin = tu.basin
		WHERE s.stop_code = ? AND cd.date = ? AND st.arrival_time >= ? AND st.arrival_time <= ?
		ORDER BY st.arrival_time;
	`, stopCode, cdDate, startTime, endTime)
	if err != nil {
		fmt.Println("GetLastStop error:", err)
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
