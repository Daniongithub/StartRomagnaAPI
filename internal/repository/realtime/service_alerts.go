package realtime

import (
	"StartRomagnaAPI/internal/model"
	"StartRomagnaAPI/internal/repository"
	"fmt"

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

func SaveAlerts(feeds map[string]*gtfs.FeedMessage) {
	for _, val := range feeds {
		for _, val2 := range val.Entity {
			fmt.Println(val2)
			fmt.Println("------")
		}
	}
}
