package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetTrips() []model.TripsResult {
	var results []model.TripsResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM trips")
	if err != nil {
		fmt.Println("GetTrips error:", err)
	}

	return results
}

// Checks if trips already exist inside the respective basins, otherwise adds them
func SaveTrips(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	trips := GetTrips()

	//Builds maps for better comparison
	tripMap := make(map[string]bool)

	for _, val := range trips {
		tripMap[val.Trip_id] = true
	}

	var new []model.TripsResult
	for _, val := range feedRA.Trips {
		_, ok := tripMap[val.Id]
		if !ok {
			var newTrip = model.ToDomainTrip(val)
			newTrip.Basin = "RA"
			new = append(new, newTrip)
		}
	}
	for _, val := range feedFC.Trips {
		_, ok := tripMap[val.Id]
		if !ok {
			var newTrip = model.ToDomainTrip(val)
			newTrip.Basin = "FC"
			new = append(new, newTrip)
		}
	}
	for _, val := range feedRN.Trips {
		_, ok := tripMap[val.Id]
		if !ok {
			var newTrip = model.ToDomainTrip(val)
			newTrip.Basin = "RN"
			new = append(new, newTrip)
		}
	}

	//Database insert
	values := make([][]any, 0, len(new))

	for _, val := range new {
		values = append(values, []any{
			val.Basin,
			val.Route_id,
			val.Service_id,
			val.Trip_id,
			val.Trip_headsign,
			val.Direction_id,
			val.Shape_id,
		})
	}

	err := BatchInsert(DB_CONTENT, "trips", []string{"basin", "route_id", "service_id", "trip_id", "trip_headsign", "direcion_id", "shape_id"}, values)
	if err != nil {
		fmt.Println("SaveTrips db error:", err)
	}
}
