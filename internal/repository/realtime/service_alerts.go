package realtime

import (
	"StartRomagnaAPI/internal/model"
	"StartRomagnaAPI/internal/repository"
	"fmt"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func GetServiceAlerts() []model.ServiceAlertsResult {
	var results []model.ServiceAlertsResult
	err := repository.DB_RT.Select(&results, "SELECT * FROM service_alerts")
	if err != nil {
		fmt.Println("GetServiceAlerts error:", err)
	}

	return results
}

func GetServiceAlertsBasin(basin string) []model.ServiceAlertsResult {
	var results []model.ServiceAlertsResult
	err := repository.DB_RT.Select(&results, "SELECT * FROM service_alerts WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetServiceAlertsBasin error:", err)
	}

	return results
}

func SaveAlerts(feeds map[string]*gtfs.FeedMessage, dbContent map[string]map[string]bool) {
	//Database insert
	for idx, val := range feeds {
		for _, val2 := range val.Entity {
			_, ok := dbContent[idx][*val2.Id]
			if !ok {
				startS := val2.Alert.ActivePeriod[0].Start
				endS := val2.Alert.ActivePeriod[0].End
				
				start := secondsToTime(int(*startS))
				end := secondsToTime(int(*endS))
				_, err := repository.DB_RT.Exec("INSERT INTO service_alerts VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", idx, val2.Id, val2.Id, start, end, val2.Alert.InformedEntity[0].RouteId, val2.Alert.InformedEntity[0].RouteType, val2.Alert.InformedEntity[0].Trip.TripId, val2.Alert.InformedEntity[0].Trip.DirectionId, val2.Alert.InformedEntity[0].Trip.StartTime, val2.Alert.InformedEntity[0].Trip.StartDate, val2.Alert.InformedEntity[0].StopId)
				if err != nil {
					fmt.Println("SaveAlerts db error:", err)
				}
			}
		}
	}
}

func secondsToTime(seconds int) time.Time {
	base := time.Now()
	midnight := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, base.Location())
	return midnight.Add(time.Duration(seconds) * time.Second)
}