package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetRoutes() []model.RoutesResult {
	var results []model.RoutesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM routes")
	if err != nil {
		fmt.Println("GetRoutes errore db:", err)
	}

	return results
}

func GetRoutesBasin(basin string) []model.RoutesResult {
	var results []model.RoutesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM routes WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetRoutes errore db:", err)
	}

	return results
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

	err := BatchInsert(DB_CONTENT, "routes", []string{"basin", "route_id", "agency_id", "route_short_name", "route_long_name", "route_type"}, values)
	if err != nil {
		fmt.Println("SaveRoutes db error:", err)
	}

	//Database delete
	for _, val := range old {
		_, err = DB_CONTENT.Exec("DELETE FROM routes WHERE basin = ? AND route_id = ?", val.Basin, val.Route_id)
		if err != nil {
			fmt.Println("SaveRoutes db error:", err)
		}
	}
}
