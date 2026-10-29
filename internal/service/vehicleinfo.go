package service

import (
	"errors"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
)

var ErrVehicleNotFound = errors.New("Vehicle not found")

func ProcessVehicleInfo(id string) (model.BusInService, error) {
	bus := realtime.GetBus(id)

	if len(bus) == 0 {
		return model.BusInService{}, ErrVehicleNotFound
	}

	newBusEnricher().enrich(&bus[0])
	return bus[0], nil
}
