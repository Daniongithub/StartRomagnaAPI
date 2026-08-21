package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"

	"github.com/Leocraft1/gtfsparser-with-reader"
)

func GetCalDates() []model.CalendarDatesResult {
	var results []model.CalendarDatesResult
	err = DB_CONTENT.Select(&results, "SELECT * FROM calendar_dates")
	if err != nil {
		fmt.Println("GetCalDates error:", err)
	}
	
	return results
}

//Same as SaveTrips
func SaveCalDates(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	dates := GetCalDates()

	datesMap := make(map[string]bool)
	for _, val := range dates {
		datesMap[val.Basin + val.Service_id+val.Date.Format("2006-01-02")] = true
	}

	var new []model.CalendarDatesResult
	collect := func(feed *gtfsparserwr.Feed, basin string) {
		for serviceId, service := range feed.Services {
			for date, added := range service.Exceptions() {
				key := basin + serviceId + date.GetTime().Format("2006-01-02")
				if _, ok := datesMap[key]; !ok {
					new = append(new, model.ToDomainException(feed.Services[serviceId], basin, date, added))
				}
			}
		}
	}

	collect(feedRA, "RA")
	collect(feedFC, "FC")
	collect(feedRN, "RN")

	//Database insert
	for _, val := range new {
		_, err := DB_CONTENT.Exec("INSERT INTO calendar_dates(basin,service_id,date,exception_type) VALUES(?, ?, ?, ?)", val.Basin, val.Service_id, val.Date, val.Exception_type)
		if err != nil {
			fmt.Println("SaveCalDates db error:", err)
		}
	}
}