package static

import (
	"fmt"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/utils"
)

func GetStopsBasin(basin string) []model.StopsResult {
	var results []model.StopsResult
	err := repository.DB_STATIC.Select(&results, `SELECT * FROM stops WHERE basin = ? AND is_dummy = 0 ORDER BY stop_code`, basin)
	if err != nil {
		fmt.Println("GetStopsBasin database error:", err)
	}

	return results
}

func GetStopsFromShapeID(basin, shapeID string) []model.StopsResult {
	var results []model.StopsResult
	err := repository.DB_STATIC.Select(&results, `
		SELECT DISTINCT s.basin, s.stop_name, s.stop_id, s.stop_code, s.stop_lat, s.stop_lon, s.is_dummy
		FROM stops s
		INNER JOIN stop_times st ON s.basin = st.basin AND s.stop_id = st.stop_id
		INNER JOIN trips t ON st.basin = t.basin AND st.trip_id = t.trip_id
		WHERE s.basin = ? AND t.shape_id = ? AND s.is_dummy = 0
		ORDER BY s.stop_code, s.stop_id
 	`, basin, shapeID)
	if err != nil {
		fmt.Println("GetStopsFromShapeID database error:", err)
	}
	return results
}

func GetLinesForStop(stopcode, basin string) []model.LineRow {
	var results []model.LineRow
	cdDate := utils.RomeNow().Format("2006-01-02")
	err := repository.DB_STATIC.Select(&results, `
        SELECT DISTINCT
            t.route_id,
            CASE 
                WHEN r.basin = 'RA' THEN r.route_short_name
                WHEN r.route_long_name IS NOT NULL AND r.route_long_name <> '' THEN r.route_long_name
                ELSE r.route_short_name 
            END AS official_line,
            h.disp_linea
        FROM stops AS s
        INNER JOIN stop_times AS st ON s.stop_id = st.stop_id AND s.basin = st.basin
        INNER JOIN trips AS t ON st.trip_id = t.trip_id AND st.basin = t.basin
        INNER JOIN routes AS r ON t.route_id = r.route_id AND t.basin = r.basin
        INNER JOIN calendar_dates AS cd ON t.service_id = cd.service_id AND t.basin = cd.basin
        LEFT JOIN headsigns AS h ON h.shape_id = t.shape_id AND h.basin = t.basin
        WHERE s.stop_code = ? AND s.basin = ? AND cd.date = ? AND cd.exception_type = 1
    `, stopcode, basin, cdDate)
	if err != nil {
		fmt.Println("GetLinesForStop database error:", err)
	}
	return results
}
