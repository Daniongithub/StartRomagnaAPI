package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/utils"
)

func ProcessRoutes(basin string) model.ClassifiedRoutes {
	routes := static.GetRoutesBasin(basin)

	//Classify routes and orders them
	return classifyRoutes(basin, routes)
}

func classifyRoutes(basin string, routes []model.RoutesResult) model.ClassifiedRoutes {
	var rows []model.RoutesWOffLine
	for _, val := range routes {
		offLine1 := static.GetOfficialLine(val.Route_id, val.Basin)
		var offLine2 string
		if offLine1 == nil {
			offLine2 = ""
		} else {
			offLine2 = *offLine1
		}
		rows = append(rows, model.RoutesWOffLine{
			Basin:          basin,
			RouteId:        val.Route_id,
			AgencyId:       val.Agency_id,
			RouteShortName: val.Route_short_name,
			RouteLongName:  val.Route_long_name,
			RouteType:      val.Route_type,
			OfficialLine:   offLine2,
		})
	}

	buckets := map[lineCategory][]model.RoutesWOffLine{
		categoryUrban:    make([]model.RoutesWOffLine, 0, len(rows)),
		categoryExurban:  make([]model.RoutesWOffLine, 0, len(rows)),
		categorySuburban: nil,
		categoryShuttle:  nil,
		categorySchool:   nil,
	}

	for _, row := range rows {
		// Separation
		category := classifyLine(basin, row.OfficialLine)
		buckets[category] = append(buckets[category], row)
	}

	for _, lines := range buckets {
		utils.SortRoutes(lines)
	}

	return model.ClassifiedRoutes{
		Urban:    buckets[categoryUrban],
		Suburban: buckets[categorySuburban],
		Exurban:  buckets[categoryExurban],
		Shuttle:  buckets[categoryShuttle],
		School:   buckets[categorySchool],
	}
}
