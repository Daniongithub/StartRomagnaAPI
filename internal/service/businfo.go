package service

import (
	"strings"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/mezzi"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
)

// Lookups live for one response only; every request reads the database.
type busEnricher struct {
	routes    map[[2]string]string
	headsigns map[[2]string]*model.Headsign
	vehicles  map[string]model.VehicleInService
}

func newBusEnricher() *busEnricher {
	return &busEnricher{make(map[[2]string]string), make(map[[2]string]*model.Headsign), make(map[string]model.VehicleInService)}
}

func (e *busEnricher) lineInfo(basin, routeID, shapeID, tripID string) (official, line, destination string) {
	routeKey, shapeKey := [2]string{basin, routeID}, [2]string{basin, shapeID}
	official, ok := e.routes[routeKey]
	if !ok {
		official = static.GetRouteNamefromID(basin, routeID)
		e.routes[routeKey] = official
	}
	headsign, ok := e.headsigns[shapeKey]
	if !ok {
		headsign = static.GetHeadsignsByID(basin, shapeID)
		e.headsigns[shapeKey] = headsign
	}
	line = official
	if headsign != nil && headsign.DispLine != nil {
		line = *headsign.DispLine
	}
	if headsign != nil && headsign.DispDest != nil {
		destination = *headsign.DispDest
	} else {
		destination = strings.ToUpper(static.GetTerminusName(basin, tripID))
	}
	return
}

func (e *busEnricher) enrich(bus *model.BusInService) {
	bus.OfficialLine, bus.Line, bus.Destination = e.lineInfo(bus.Basin, bus.RouteId, bus.ShapeId, bus.TripId)
	bus.NextStop = realtime.GetFirstStop(bus.TripId)
	info, ok := e.vehicles[bus.Vehicle]
	if !ok {
		vehicle := mezzi.GetVehicleInServiceByID(bus.Vehicle)
		if vehicle == nil {
			vehicle = mezzi.GetMeteInServiceByID(bus.Vehicle)
		}
		info = model.VehicleInService{Number: bus.Vehicle}
		if vehicle != nil {
			info = *vehicle
		}
		e.vehicles[bus.Vehicle] = info
	}
	bus.VehicleInfo = info
}
