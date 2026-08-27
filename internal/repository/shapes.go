package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func GetShapes() []model.ShapesResult {
	var results []model.ShapesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM shapes")
	if err != nil {
		fmt.Println("GetShapes errore db:", err)
	}

	return results
}

func GetShapesBasin(basin string) []model.ShapesResult {
	var results []model.ShapesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM shapes WHERE basin = ?")
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
	for idx, val := range feedRA.Shapes {
		_, ok := shapesMap["RA"+idx]
		if !ok {
			for _, val2 := range val.Points {
				newShape := model.ToDomainShapes("RA", val.Id, val2.Lat, val2.Lon, int(val2.Sequence))
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedFC.Shapes {
		_, ok := shapesMap["FC"+idx]
		if !ok {
			for _, val2 := range val.Points {
				newShape := model.ToDomainShapes("FC", val.Id, val2.Lat, val2.Lon, int(val2.Sequence))
				new = append(new, newShape)
			}
		}
	}
	for idx, val := range feedRN.Shapes {
		_, ok := shapesMap["RN"+idx]
		if !ok {
			for _, val2 := range val.Points {
				newShape := model.ToDomainShapes("RN", val.Id, val2.Lat, val2.Lon, int(val2.Sequence))
				new = append(new, newShape)
			}
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

	err := BatchInsert(DB_CONTENT, "shapes", []string{"basin", "shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence"}, values)
	if err != nil {
		fmt.Println("SaveShapes db error:", err)
	}
}
