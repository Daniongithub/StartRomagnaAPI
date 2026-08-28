package static

import (
	"StartRomagnaAPI/internal/model"
	"StartRomagnaAPI/internal/repository"
	"fmt"
	"time"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetCalDates() []model.CalendarDatesResult {
	var results []model.CalendarDatesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM calendar_dates")
	if err != nil {
		fmt.Println("GetCalDates error:", err)
	}

	return results
}

func GetCalDatesBasin(basin string) []model.CalendarDatesResult {
	var results []model.CalendarDatesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM calendar_dates WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetCalDates error:", err)
	}

	return results
}

func GetCalDatesRange() (string, string) {
	var min, max time.Time

	err := repository.DB_STATIC.QueryRow("SELECT MIN(date), MAX(date) FROM calendar_dates").Scan(&min, &max)

	if err != nil {
		fmt.Println("GetCalDatesRange error:", err)
		return "", ""
	}

	return min.Format("2006-01-02"), max.Format("2006-01-02")
}

// Same as SaveTrips
func SaveCalDates(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	dates := GetCalDates()
	datesMap := make(map[string]bool)
	for _, val := range dates {
		datesMap[val.Basin+val.Service_id+val.Date.Format("2006-01-02")] = true
	}

	var new []model.CalendarDatesResult
	var old []model.CalendarDatesResult

	//Track every row in gtfs
	feedKeys := make(map[string]bool)

	collect := func(feed *gtfsparserwr.Feed, basin string) {
		for serviceId, service := range feed.Services {
			for date, added := range service.Exceptions() {
				key := basin + serviceId + date.GetTime().Format("2006-01-02")
				feedKeys[key] = true

				if _, ok := datesMap[key]; !ok {
					new = append(new, model.ToDomainException(feed.Services[serviceId], basin, date, added))
				}
			}
		}
	}

	collect(feedRA, "RA")
	collect(feedFC, "FC")
	collect(feedRN, "RN")

	//Removes old entries
	for _, val := range dates {
		key := val.Basin + val.Service_id + val.Date.Format("2006-01-02")
		if _, ok := feedKeys[key]; !ok {
			var row model.CalendarDatesResult
			row.Basin = val.Basin
			row.Service_id = val.Service_id
			row.Date = val.Date
			old = append(old, row)
		}
	}

	//Database insert
	values := make([][]any, 0, len(new))

	for _, val := range new {
		values = append(values, []any{
			val.Basin,
			val.Service_id,
			val.Date,
			val.Exception_type,
		})
	}

	err := repository.BatchInsert(repository.DB_STATIC, "calendar_dates", []string{"basin", "service_id", "date", "exception_type"}, values)
	if err != nil {
		fmt.Println("SaveCalDates db error:", err)
	}

	//Database delete
	for _, val := range old {
		_, err = repository.DB_STATIC.Exec("DELETE FROM calendar_dates WHERE basin = ? AND service_id = ? AND date = ?", val.Basin, val.Service_id, val.Date)
		if err != nil {
			fmt.Println("SaveCalDates db error:", err)
		}
	}
}
