package static

import (
	"fmt"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
)

func GetRoutes() []model.RoutesResult {
	var results []model.RoutesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM routes")
	if err != nil {
		fmt.Println("GetRoutes database error:", err)
	}

	return results
}

func GetRoutesBasin(basin string) []model.RoutesResult {
	var results []model.RoutesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM routes WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetRoutesBasin database error:", err)
	}

	return results
}

func GetRouteNamefromID(basin, routeId string) string {
	var result []string
	err := repository.DB_STATIC.Select(&result, "SELECT CASE WHEN basin = 'RA' THEN route_short_name WHEN route_long_name IS NOT NULL AND route_long_name <> '' THEN route_long_name ELSE route_short_name END FROM routes WHERE route_id = ? AND basin = ? LIMIT 1", routeId, basin)
	if err != nil {
		fmt.Println("GetRouteNamefromId database error:", err)
	}

	if len(result) == 0 {
		return ""
	}
	return result[0]
}
