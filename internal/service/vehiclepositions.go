package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/mezzi"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"strings"
)

func ProcessVehiclePositions() []model.VehiclePosition {
	positions := realtime.GetVehiclePositions()

	for idx := range positions {
		val := &positions[idx]
		val.OfficialLine = static.GetRouteNamefromID(val.Basin, val.RouteId)
		headsign := static.GetHeadsignsByID(val.ShapeId)
		if headsign != nil {
			if headsign.DispLine != nil {
				val.Line = *headsign.DispLine
			} else {
				val.Line = val.OfficialLine
			}
			if headsign.DispDest != nil {
				val.Destination = *headsign.DispDest
			} else {
				val.Destination = strings.ToUpper(static.GetTerminusName(val.TripId))
			}
		}
		if mezzi.GetVehicleInServiceByID(val.Vehicle) != nil {
			val.VehicleInfo = *mezzi.GetVehicleInServiceByID(val.Vehicle)
		} else {
			val.VehicleInfo = model.VehicleInService{
				Number: val.Vehicle,
			}
		}
	}

	return positions
}

func ProcessVehiclePositionsBasin(basin string) []model.VehiclePosition {
	positions := realtime.GetVehiclePositionsBasin(basin)

	for idx := range positions {
		val := &positions[idx]
		val.OfficialLine = static.GetRouteNamefromID(val.Basin, val.RouteId)
		headsign := static.GetHeadsignsByID(val.ShapeId)
		if headsign != nil {
			if headsign.DispLine != nil {
				val.Line = *headsign.DispLine
			} else {
				val.Line = val.OfficialLine
			}
			if headsign.DispDest != nil {
				val.Destination = *headsign.DispDest
			} else {
				val.Destination = strings.ToUpper(static.GetTerminusName(val.TripId))
			}
		}
		if mezzi.GetVehicleInServiceByID(val.Vehicle) != nil {
			val.VehicleInfo = *mezzi.GetVehicleInServiceByID(val.Vehicle)
		} else {
			val.VehicleInfo = model.VehicleInService{
				Number: val.Vehicle,
			}
		}
	}

	return positions
}