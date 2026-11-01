package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/utils"
)

func ProcessRoutes(basin string) []model.RoutesResult {
	routes := static.GetRoutesBasin(basin)

	var lines []model.Line
	//Parses to Line format
	for _, val := range routes {
		lines = append(lines, model.Line{
			Line: val.Route_id,
		})
	}
	//Calls to existing function
	utils.SortLines(lines)

	var sorted []model.RoutesResult
	//Reconstructs ordered array
	for _, val := range lines {
		original := utils.FindRouteFromID(routes, val.Line);
		if original == nil {
			return nil
		}
		sorted = append(sorted, model.RoutesResult{
			Basin:    basin,
			Route_id: val.Line,
			Agency_id: original.Agency_id,
			Route_short_name: original.Route_short_name,
			Route_long_name: original.Route_long_name,
			Route_type: original.Route_type,
		})
	}
	return sorted
}
