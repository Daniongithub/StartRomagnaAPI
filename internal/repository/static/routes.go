package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetRoutes() []model.RoutesResult {
	var results []model.RoutesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM routes")
	if err != nil {
		fmt.Println("GetRoutes errore db:", err)
	}

	return results
}

func GetRoutesBasin(basin string) []model.RoutesResult {
	var results []model.RoutesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM routes WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetRoutesBasin errore db:", err)
	}

	return results
}

func GetRouteName(basin string) string {
	var result []string
	err := repository.DB_STATIC.Select(&result, "SELECT CASE WHEN basin = 'RA' THEN route_short_name WHEN route_long_name IS NOT NULL AND route_long_name <> '' THEN route_long_name ELSE route_short_name END FROM routes WHERE basin = ? LIMIT 1", basin)
	if err != nil {
		fmt.Println("GetRouteName errore db:", err)
	}

	return result[0]
}

func GetRouteNamefromID(basin, routeId string) string {
	var result []string
	err := repository.DB_STATIC.Select(&result, "SELECT CASE WHEN basin = 'RA' THEN route_short_name WHEN route_long_name IS NOT NULL AND route_long_name <> '' THEN route_long_name ELSE route_short_name END FROM routes WHERE route_id = ? AND basin = ? LIMIT 1", routeId, basin)
	if err != nil {
		fmt.Println("GetRouteNamefromId errore db:", err)
	}

	return result[0]
}

func SaveRoutes(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	routes := GetRoutes()

	routesMap := make(map[string]bool)
	for _, val := range routes {
		routesMap[val.Basin+val.Route_id] = true
	}

	var new []model.RoutesResult
	var old []model.RoutesResult
	feedKeys := make(map[string]bool)
	for idx, val := range feedRA.Routes {
		_, ok := routesMap["RA"+idx]
		feedKeys["RA"+idx] = true
		if !ok {
			newRoute := model.ToDomainRoutes(val)
			newRoute.Basin = "RA"
			new = append(new, newRoute)
		}
	}

	for idx, val := range feedFC.Routes {
		_, ok := routesMap["FC"+idx]
		feedKeys["FC"+idx] = true
		if !ok {
			newRoute := model.ToDomainRoutes(val)
			newRoute.Basin = "FC"
			new = append(new, newRoute)
		}
	}
	for idx, val := range feedRN.Routes {
		_, ok := routesMap["RN"+idx]
		feedKeys["RN"+idx] = true
		if !ok {
			newRoute := model.ToDomainRoutes(val)
			newRoute.Basin = "RN"
			new = append(new, newRoute)
		}
	}

	for _, val := range routes {
		_, ok := feedKeys[val.Basin+val.Route_id]
		if !ok {
			var oldRoute model.RoutesResult
			oldRoute.Basin = val.Basin
			oldRoute.Route_id = val.Route_id
			old = append(old, oldRoute)
		}
	}

	//Database insert
	values := make([][]any, 0, len(new))

	for _, val := range new {
		values = append(values, []any{
			val.Basin,
			val.Route_id,
			val.Agency_id,
			val.Route_short_name,
			val.Route_long_name,
			val.Route_type,
		})
	}

	err := repository.BatchInsert(repository.DB_STATIC, "routes", []string{"basin", "route_id", "agency_id", "route_short_name", "route_long_name", "route_type"}, values)
	if err != nil {
		fmt.Println("SaveRoutes db error:", err)
	}

	//Database delete
	for _, val := range old {
		_, err = repository.DB_STATIC.Exec("DELETE FROM routes WHERE basin = ? AND route_id = ?", val.Basin, val.Route_id)
		if err != nil {
			fmt.Println("SaveRoutes db error:", err)
		}
	}
}
