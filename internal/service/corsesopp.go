package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
)

func ProcessCorsesopp(basin string) []model.CorseSopp {
	serviceAlerts := realtime.GetDistinctSAByBasin(basin)

	var corseSopp []model.CorseSopp
	for _, val := range serviceAlerts {
		firstStop := realtime.GetFirstStop(val.Trip_id)
		lastStop := realtime.GetLastStop(val.Trip_id)
		vehicle := realtime.GetVehicleByTripId(val.Trip_id)
		corsa := model.CorseSopp{
			RouteId: val.Route_id,
			Start: val.Start.Format("15:04:05"),
			End: val.End.Format("15:04:05"),
			StartDate: val.Start_date.Format("02-01-2006"),
			TripId: val.Trip_id,
			DirectionId: val.Direction_id,
			FirstStop: firstStop,
			LastStop: lastStop,
			Vehicle: checkVehicle(vehicle),
		}
		corseSopp = append(corseSopp, corsa)
	}

	return corseSopp
}

func checkVehicle(vehicle string) *string {
	if vehicle == "0" {
		return nil
	}
	return &vehicle
}
