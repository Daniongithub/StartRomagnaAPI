package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
)

func ProcessShape(basin, shapeId string) model.Shape {
	var shape model.Shape
	shape.Points = static.GetShapePoints(basin, shapeId)
	shape.Stops = static.GetStopsFromShapeID(basin, shapeId)

	return shape
}
