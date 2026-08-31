package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"

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
	err := repository.DB_STATIC.Select(&results, `SELECT * FROM stops WHERE LOCATE('semaforo', stop_name) = 0
      AND LOCATE('fi1', stop_name) = 0
	  AND LOCATE('FITTIZIO', stop_name) = 0
      AND LOCATE('Fittizio', stop_name) = 0
      AND LOCATE('FITTIZIA', stop_name) = 0
      AND LOCATE('Fittizia', stop_name) = 0`)
	if err != nil {
		fmt.Println("GetStopsFiltered errore db:", err)
	}

	return results
}

func GetStopsBasin(basin string) []model.StopsResult {
	var results []model.StopsResult
	err := repository.DB_STATIC.Select(&results, `SELECT * FROM stops WHERE basin = ? AND LOCATE('semaforo', stop_name) = 0
	  AND LOCATE('fi1', stop_name) = 0
      AND LOCATE('FITTIZIO', stop_name) = 0
      AND LOCATE('Fittizio', stop_name) = 0
      AND LOCATE('FITTIZIA', stop_name) = 0
      AND LOCATE('Fittizia', stop_name) = 0
	  ORDER BY stop_code`, basin)
	if err != nil {
		fmt.Println("GetStopsBasin errore db:", err)
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
}
