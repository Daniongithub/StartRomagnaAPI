package static

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/notify"
	"startromagnaapi/internal/repository"
)

func GetShapesBasin(basin string) []model.ShapesResult {
	var results []model.ShapesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM shapes WHERE basin = ?", basin)
	if err != nil {
		notify.ReportError("Failed to read shapes by basin", err)
	}

	return results
}

func GetShapePoints(basin, shapeId string) []model.ShapesResult {
	var results []model.ShapesResult
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM shapes WHERE basin = ? AND shape_id = ? ORDER BY shape_pt_sequence", basin, shapeId)
	if err != nil {
		notify.ReportError("Failed to read shape points", err)
	}

	return results
}
