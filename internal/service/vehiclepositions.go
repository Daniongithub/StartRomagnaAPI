package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/mezzi"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"strings"
)

func ProcessVehiclePositions() []model.BusInService {
	positions := realtime.GetVehiclePositions()

	for idx := range positions {
		val := &positions[idx]
		val.OfficialLine = static.GetRouteNamefromID(val.Basin, val.RouteId)
		val.NextStop = realtime.GetFirstStop(val.TripId)
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

func ProcessVehiclePositionsBasin(basin string) []model.BusInService {
	positions := realtime.GetVehiclePositionsBasin(basin)

	for idx := range positions {
		val := &positions[idx]
		val.OfficialLine = static.GetRouteNamefromID(val.Basin, val.RouteId)
		val.NextStop = realtime.GetFirstStop(val.TripId)
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

func ProcessVehiclePositionID(id string) *model.BusInService {
	position := realtime.GetVehicleLocationID(id)

	if position == nil {
		return nil
	}

	position.OfficialLine = static.GetRouteNamefromID(position.Basin, position.RouteId)
	position.NextStop = realtime.GetFirstStop(position.TripId)
	headsign := static.GetHeadsignsByID(position.ShapeId)
	if headsign != nil {
		if headsign.DispLine != nil {
			position.Line = *headsign.DispLine
		} else {
			position.Line = position.OfficialLine
		}
		if headsign.DispDest != nil {
			position.Destination = *headsign.DispDest
		} else {
			position.Destination = strings.ToUpper(static.GetTerminusName(position.TripId))
		}
	}
	if mezzi.GetVehicleInServiceByID(position.Vehicle) != nil {
		position.VehicleInfo = *mezzi.GetVehicleInServiceByID(position.Vehicle)
	} else {
		position.VehicleInfo = model.VehicleInService{
			Number: position.Vehicle,
		}
	}

	return position
}
