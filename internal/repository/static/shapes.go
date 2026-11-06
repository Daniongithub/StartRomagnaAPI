package static

import (
	"fmt"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
)

func GetShapesBasin(basin string) []model.ShapesResult {
	var results []model.ShapesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM shapes WHERE basin = ?", basin)
	if err != nil {
		fmt.Println("GetShapesBasin database error:", err)
	}

	return results
}

func GetShapePoints(basin, shapeId string) []model.ShapesResult {
	var results []model.ShapesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM shapes WHERE basin = ? AND shape_id = ? ORDER BY shape_pt_sequence", basin, shapeId)
	if err != nil {
		fmt.Println("GetShapePoints database error:", err)
	}

	return results
}
