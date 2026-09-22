package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"time"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetStops() []model.StopsResult {
	var results []model.StopsResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM stops")
	if err != nil {
		fmt.Println("GetStops errore db:", err)
	}

	return results
}

func GetStopsFiltered() []model.StopsResult {
	var results []model.StopsResult
	err := repository.DB_STATIC.Select(&results, `SELECT * FROM stops WHERE is_dummy = 0`)
	if err != nil {
		fmt.Println("GetStopsFiltered errore db:", err)
	}

	return results
}

func GetStopsBasin(basin string) []model.StopsResult {
	var results []model.StopsResult
	err := repository.DB_STATIC.Select(&results, `SELECT * FROM stops WHERE basin = ? AND is_dummy = 0 ORDER BY stop_code`, basin)
	if err != nil {
		fmt.Println("GetStopsBasin errore db:", err)
	}

	return results
}

func GetStopsFromTripID(basin, tripId string) []model.StopsResult {
	var results []model.StopsResult
	err := repository.DB_STATIC.Select(&results, `
		SELECT s.basin, s.stop_name, s.stop_id, s.stop_code, s.stop_lat, s.stop_lon, s.is_dummy FROM stops AS s
		INNER JOIN stop_times AS st
		ON s.stop_id = st.stop_id AND s.basin = st.basin
		WHERE s.basin = ? AND st.trip_id = ? AND s.is_dummy = 0 
	`, basin, tripId)
	if err != nil {
		fmt.Println("GetStopsFromTripID errore db:", err)
	}

	return results
}

func GetLinesForStop(stopcode, basin string) []model.LineRow {
    var results []model.LineRow
	now := time.Now()
	cdDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
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
        LEFT JOIN headsigns AS h ON h.shape_id = t.shape_id
        WHERE s.stop_code = ? AND s.basin = ? AND cd.date = ?
    `, stopcode, basin, cdDate)
    if err != nil {
        fmt.Println("GetLinesForStop errore db:", err)
    }
    return results
}

func SaveStops(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	stops := GetStops()

	stopMap := make(map[string]bool)
	for _, val := range stops {
		stopMap[val.Basin+val.Stop_id] = true
	}

	var new []model.StopsResult
	var old []model.StopsResult
	feedKeys := make(map[string]bool)
	for idx, val := range feedRA.Stops {
		_, ok := stopMap["RA"+idx]
		feedKeys["RA"+idx] = true
		if !ok {
			newShape := model.ToDomainStops(val)
			newShape.Basin = "RA"
			new = append(new, newShape)
		}
	}
	for idx, val := range feedFC.Stops {
		_, ok := stopMap["FC"+idx]
		feedKeys["FC"+idx] = true
		if !ok {
			newShape := model.ToDomainStops(val)
			newShape.Basin = "FC"
			new = append(new, newShape)
		}
	}

	for idx, val := range feedRN.Stops {
		_, ok := stopMap["RN"+idx]
		feedKeys["RN"+idx] = true
		if !ok {
			newShape := model.ToDomainStops(val)
			newShape.Basin = "RN"
			new = append(new, newShape)
		}
	}

	for _, val := range stops {
		_, ok := feedKeys[val.Basin+val.Stop_id]
		if !ok {
			var oldStop model.StopsResult
			oldStop.Basin = val.Basin
			oldStop.Stop_id = val.Stop_id
			old = append(old, oldStop)
		}
	}

	//Database insert
	values := make([][]any, 0, len(new))

	for _, val := range new {
		values = append(values, []any{
			val.Basin,
			val.Stop_id,
			val.Stop_code,
			val.Stop_name,
			val.Stop_lat,
			val.Stop_lon,
		})
	}

	err := repository.BatchInsert(repository.DB_STATIC, "stops", []string{"basin", "stop_id", "stop_code", "stop_name", "stop_lat", "stop_lon"}, values)
	if err != nil {
		fmt.Println("SaveStops db error:", err)
	}

	//Database delete
	for _, val := range old {
		_, err = repository.DB_STATIC.Exec("DELETE FROM stops WHERE basin = ? AND stop_id = ?", val.Basin, val.Stop_id)
		if err != nil {
			fmt.Println("SaveStops db error:", err)
		}
	}

	//Sets fake stops
	_, err = repository.DB_STATIC.Exec(`
		UPDATE start_gtfs_static.stops
		SET is_dummy = 1
		WHERE LOCATE('semaforo', stop_name) > 0
		OR LOCATE('fi1', stop_name) > 0
		OR LOCATE('FITTIZIO', stop_name) > 0
		OR LOCATE('Fittizio', stop_name) > 0
		OR LOCATE('FITTIZIA', stop_name) > 0
		OR LOCATE('Fittizia', stop_name) > 0
	`)
	if err != nil {
		fmt.Println("SaveStops db error:", err)
	}
}
