package static

import (
	"fmt"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/utils"
)

func GetTripsBasin(basin string) []model.TripsResult {
	var results []model.TripsResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM trips WHERE basin = ? ORDER BY route_id", basin)
	if err != nil {
		fmt.Println("GetTripsBasin error:", err)
	}

	return results
}

func GetTodayTripInfoFromRouteID(basin, routeId, directionId string) []model.TripInfo {
	var results []model.TripInfo
	cdDate := utils.RomeNow().Format("2006-01-02")
	err := repository.DB_STATIC.Select(&results, `
		SELECT t.basin, t.trip_id, t.shape_id FROM trips AS t
		INNER JOIN calendar_dates AS cd
		ON t.service_id = cd.service_id AND t.basin = cd.basin
		LEFT JOIN stop_times AS st
		ON t.trip_id = st.trip_id AND t.basin = st.basin
		WHERE t.basin = ? AND t.route_id = ? AND cd.date = ? AND cd.exception_type = 1 AND t.direction_id = ? AND st.stop_sequence = (
      		SELECT MIN(st2.stop_sequence)
      		FROM stop_times AS st2
      		WHERE st2.trip_id = t.trip_id
        	AND st2.basin = t.basin
		)
		ORDER BY st.arrival_time
	`, basin, routeId, cdDate, directionId)
	if err != nil {
		fmt.Println("GetTripIDsFromRouteID error:", err)
	}

	return results
}
