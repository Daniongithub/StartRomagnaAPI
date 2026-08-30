package static

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"fmt"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetShapes() []model.ShapesResult {
	var results []model.ShapesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM shapes")
	if err != nil {
		fmt.Println("GetShapes errore db:", err)
	}

	return results
}

func GetShapesBasin(basin string) []model.ShapesResult {
	var results []model.ShapesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM shapes WHERE basin = ?")
	if err != nil {
		fmt.Println("GetShapes errore db:", err)
	}

	return results
}

func SaveShapes(feedRA *gtfsparserwr.Feed, feedFC *gtfsparserwr.Feed, feedRN *gtfsparserwr.Feed) {
	shapes := GetShapes()

	shapesMap := make(map[string]bool)
	for _, val := range shapes {
		shapesMap[val.Basin+val.Shape_id] = true
	}

	var new []model.ShapesResult
	var old []model.ShapesResult
	feedKeys := make(map[string]bool)
	for idx, val := range feedRA.Shapes {
		_, ok := shapesMap["RA"+idx]
		feedKeys["RA"+idx] = true
		if !ok {
			for _, val2 := range val.Points {
				newShape := model.ToDomainShapes("RA", val.Id, val2.Lat, val2.Lon, int(val2.Sequence))
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedFC.Shapes {
		_, ok := shapesMap["FC"+idx]
		feedKeys["FC"+idx] = true
		if !ok {
			for _, val2 := range val.Points {
				newShape := model.ToDomainShapes("FC", val.Id, val2.Lat, val2.Lon, int(val2.Sequence))
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedRN.Shapes {
		_, ok := shapesMap["RN"+idx]
		feedKeys["RN"+idx] = true
		if !ok {
			for _, val2 := range val.Points {
				newShape := model.ToDomainShapes("RN", val.Id, val2.Lat, val2.Lon, int(val2.Sequence))
				new = append(new, newShape)
			}
		}
	}

	for _, val := range shapes {
		_, ok := feedKeys[val.Basin+val.Shape_id]
		if !ok {
			var oldShape model.ShapesResult
			oldShape.Basin = val.Basin
			oldShape.Shape_id = val.Shape_id
			oldShape.Shape_pt_sequence = val.Shape_pt_sequence
			old = append(old, oldShape)
		}
	}

	//Database insert
	values := make([][]any, 0, len(new))

	for _, val := range new {
		values = append(values, []any{
			val.Basin,
			val.Shape_id,
			val.Shape_pt_lat,
			val.Shape_pt_lon,
			val.Shape_pt_sequence,
		})
	}

	err := repository.BatchInsert(repository.DB_STATIC, "shapes", []string{"basin", "shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence"}, values)
	if err != nil {
		fmt.Println("SaveShapes db error:", err)
	}

	//Database delete
	for _, val := range old {
		_, err = repository.DB_STATIC.Exec("DELETE FROM shapes WHERE basin = ? AND shape_id = ? AND shape_pt_sequence = ?", val.Basin, val.Shape_id, val.Shape_pt_sequence)
		if err != nil {
			fmt.Println("SaveStops db error:", err)
		}
	}
}
