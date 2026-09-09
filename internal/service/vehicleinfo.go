package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/mezzi"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"strings"
)

func ProcessVehicleInfo(id string) model.BusInService {
	bus := realtime.GetBus(id)

	val := &bus[0]
	val.NextStop = realtime.GetFirstStop(val.TripId)
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
	//Model and stuff
	if mezzi.GetVehicleInServiceByID(val.Vehicle) != nil {
		val.VehicleInfo = *mezzi.GetVehicleInServiceByID(val.Vehicle)
	} else if mezzi.GetMeteInServiceByID(val.Vehicle) != nil {
		//METE model and stuff
		val.VehicleInfo = *mezzi.GetMeteInServiceByID(val.Vehicle)
	} else {
		val.VehicleInfo = model.VehicleInService{
			Number: val.Vehicle,
		}
	}

	return bus[0]
}
