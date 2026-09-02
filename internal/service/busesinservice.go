package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
)

func ProcessBusesInService() []model.BusInService {
	buses := realtime.GetBuses()

	for idx := range buses {
		val := &buses[idx]
		val.NextStop = realtime.GetFirstStop(val.TripId)
		val.OfficialLine = static.GetRouteNamefromId(val.Basin, val.RouteId)
	}

	return buses
}
