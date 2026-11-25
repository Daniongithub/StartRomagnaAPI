package static

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/notify"
	"startromagnaapi/internal/repository"
)

func GetHeadsignsByID(basin, shapeId string) *model.Headsign {
	var results []model.Headsign
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM headsigns WHERE shape_id = ? AND basin = ?", shapeId, basin)
	if err != nil {
		notify.ReportError("Failed to read headsigns by ID", err)
	}
	if len(results) == 0 {
		return nil
	}

	return &results[0]
}

func GetHeadsignsByRouteID(basin, routeid string) []model.ShapeIDWInfo {
	var results []model.ShapeIDWInfo
	err := repository.DB_STATIC.Select(&results, `
		SELECT h.shape_id, h.disp_dest,
		CASE 
			WHEN h.disp_linea IS NOT NULL THEN h.disp_linea
			WHEN r.disp_linea IS NOT NULL THEN r.disp_linea
			WHEN r.basin = 'RA' THEN r.route_short_name
			WHEN r.route_long_name IS NOT NULL AND r.route_long_name <> '' THEN r.route_long_name
			ELSE r.route_short_name
		END AS disp_linea
		FROM headsigns AS h
		INNER JOIN trips AS t
		ON h.shape_id = t.shape_id AND h.basin = t.basin
		INNER JOIN routes AS r
		ON r.route_id = t.route_id AND r.basin = t.basin
		WHERE h.basin = ? AND t.route_id = ?
	`, basin, routeid)
	if err != nil {
		notify.ReportError("Failed to read a route's shapes", err)
	}

	return results
}
