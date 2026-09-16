package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetTrips() []model.TripsResult {
	var results []model.TripsResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM trips")
	if err != nil {
		fmt.Println("GetTrips error:", err)
	}

	return results
}

func GetTripsBasin(basin string) []model.TripsResult {
	var results []model.TripsResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM trips WHERE basin = ? ORDER BY route_id", basin)
	if err != nil {
		fmt.Println("GetTripsBasin error:", err)
	}

	return results
}

func GetRouteIDFromShape(basin, shapeid string) string {
	var results []string
	err := repository.DB_STATIC.Select(&results, `
		SELECT t.route_id FROM trips AS t
		WHERE t.shape_id = ? AND t.basin = ?
	`, shapeid, basin)
	if err != nil {
		fmt.Println("GetRouteIDsFromShape errore db:", err)
	}

	return results[0]
}

// Checks if trips already exist inside the respective basins, otherwise adds them
func SaveTrips(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	values := make([][]any, 0, len(feedRA.Trips)+len(feedFC.Trips)+len(feedRN.Trips))

	for _, val := range feedRA.Trips {
		trip := model.ToDomainTrip(val)
		trip.Basin = "RA"

		values = append(values, []any{
			trip.Basin,
			trip.Route_id,
			trip.Service_id,
			trip.Trip_id,
			trip.Trip_headsign,
			trip.Direction_id,
			trip.Shape_id,
		})
	}

	for _, val := range feedFC.Trips {
		trip := model.ToDomainTrip(val)
		trip.Basin = "FC"

		values = append(values, []any{
			trip.Basin,
			trip.Route_id,
			trip.Service_id,
			trip.Trip_id,
			trip.Trip_headsign,
			trip.Direction_id,
			trip.Shape_id,
		})
	}

	for _, val := range feedRN.Trips {
		trip := model.ToDomainTrip(val)
		trip.Basin = "RN"

		values = append(values, []any{
			trip.Basin,
			trip.Route_id,
			trip.Service_id,
			trip.Trip_id,
			trip.Trip_headsign,
			trip.Direction_id,
			trip.Shape_id,
		})
	}

	// Svuota la tabella
	_, err := repository.DB_STATIC.Exec("DELETE FROM trips")
	if err != nil {
		fmt.Println("SaveTrips delete error:", err)
		return
	}

	// Reinserisce tutti i trip
	err = repository.BatchInsert(repository.DB_STATIC, "trips", []string{"basin", "route_id", "service_id", "trip_id", "trip_headsign", "direction_id", "shape_id"}, values)
	if err != nil {
		fmt.Println("SaveTrips db error:", err)
	}
}
