package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
)

func ProcessShape(basin, shapeId string) model.Shape {
	var shape model.Shape
	shape.Points = static.GetShapePoints(basin, shapeId)
	tripIDs := static.GetTripIDsFromShape(basin, shapeId)
	seenStops := make(map[string]model.StopsResult)
	for _, val := range tripIDs {
		stops := static.GetStopsFromTripID(basin, val)
		for _, val2 := range stops {
			seenStops[val2.Basin+val2.Stop_code] = val2
		}
	}
	for _, val := range seenStops {
		shape.Stops = append(shape.Stops, val)
	}

	return shape
}