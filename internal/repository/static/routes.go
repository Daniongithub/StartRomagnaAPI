package static

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/notify"
	"startromagnaapi/internal/repository"
)

func GetRoutes() []model.RoutesResult {
	var results []model.RoutesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM routes")
	if err != nil {
		notify.ReportError("Failed to read routes", err)
	}

	return results
}

func GetRoutesBasin(basin string) []model.RoutesResult {
	var results []model.RoutesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM routes WHERE basin = ?", basin)
	if err != nil {
		notify.ReportError("Failed to read routes by basin", err)
	}

	return results
}

func GetRouteNamefromID(basin, routeId string) string {
	var result []string
	err := repository.DB_STATIC.Select(&result, "SELECT CASE WHEN basin = 'RA' THEN route_short_name WHEN route_long_name IS NOT NULL AND route_long_name <> '' THEN route_long_name ELSE route_short_name END FROM routes WHERE route_id = ? AND basin = ? LIMIT 1", routeId, basin)
	if err != nil {
		notify.ReportError("Failed to read a route name", err)
	}

	if len(result) == 0 {
		return ""
	}
	return result[0]
}

func GetOfficialLine(routeId, basin string) *string {
	var results []string
	err := repository.DB_STATIC.Select(&results, `
        SELECT DISTINCT
            CASE 
                WHEN r.basin = 'RA' THEN r.route_short_name
                WHEN r.route_long_name IS NOT NULL AND r.route_long_name <> '' THEN r.route_long_name
                ELSE r.route_short_name 
            END AS official_line
        FROM routes AS r
        WHERE r.route_id = ? AND r.basin = ?
    `, routeId, basin)
	if err != nil {
		notify.ReportError("Failed to read lines by stop", err)
	}
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

func GetLineName(routeId, basin string) *string {
	var results []string
	err := repository.DB_STATIC.Select(&results, `
        SELECT DISTINCT
            CASE 
				WHEN r.disp_linea IS NOT NULL THEN r.disp_linea
                WHEN r.basin = 'RA' THEN r.route_short_name
                WHEN r.route_long_name IS NOT NULL AND r.route_long_name <> '' THEN r.route_long_name
				ELSE r.route_short_name
            END AS official_line
        FROM routes AS r
        WHERE r.route_id = ? AND r.basin = ?
    `, routeId, basin)
	if err != nil {
		notify.ReportError("Failed to read the official line name", err)
	}
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}
