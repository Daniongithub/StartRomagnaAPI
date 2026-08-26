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

func SaveRoutes(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	routes := GetRoutes()

	routesMap := make(map[string]bool)
	for _, val := range routes {
		routesMap[val.Basin+val.Route_id] = true
	}

	var new []model.RoutesResult
	for idx, val := range feedRA.Routes {
		_, ok := routesMap["RA"+idx]
		if !ok {
			newRoute := model.ToDomainRoutes(val)
			newRoute.Basin = "RA"
			new = append(new, newRoute)
		}
	}
	for idx, val := range feedFC.Routes {
		_, ok := routesMap["FC"+idx]
		if !ok {
			newRoute := model.ToDomainRoutes(val)
			newRoute.Basin = "FC"
			new = append(new, newRoute)
		}
	}
	for idx, val := range feedRN.Routes {
		_, ok := routesMap["RN"+idx]
		if !ok {
			newRoute := model.ToDomainRoutes(val)
			newRoute.Basin = "RN"
			new = append(new, newRoute)
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
}
