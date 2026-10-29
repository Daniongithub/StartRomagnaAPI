package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
)

func ProcessVehiclePositions() []model.BusInService {
	positions := realtime.GetVehiclePositions()
	enrich := newBusEnricher()
	for i := range positions {
		enrich.enrich(&positions[i])
	}
	return positions
}

func ProcessVehiclePositionsBasin(basin string) []model.BusInService {
	positions := realtime.GetVehiclePositionsBasin(basin)
	enrich := newBusEnricher()
	for i := range positions {
		enrich.enrich(&positions[i])
	}
	return positions
}

func ProcessVehiclePositionID(id string) *model.BusInService {
	position := realtime.GetVehicleLocationID(id)
	if position != nil {
		newBusEnricher().enrich(position)
	}
	return position
}
