package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"strings"
)

func ProcessTimetable(basin, routeId, directionId string) []model.Trip {
	trips := static.GetTodayTripInfoFromRouteID(basin, routeId, directionId)
	var timetable []model.Trip
	enrich := newBusEnricher()
	for _, val := range trips {
		val.OfficialLine, val.Line, val.Destination = enrich.lineInfo(val.Basin, routeId, val.ShapeId, val.TripId)
		// Format arrival and departure times.
		formattedTimes := static.GetStopTimesDetailByTrip(basin, val.TripId)
		for idx := range formattedTimes {
			val := &formattedTimes[idx]
			arrival := val.ArrivalTime.SQLString()
			val.ArrivalTimeStr = arrival[:strings.LastIndex(arrival, ":")]
			departure := val.DepartureTime.SQLString()
			val.DepartureTimeStr = departure[:strings.LastIndex(departure, ":")]
		}
		timetable = append(timetable, model.Trip{
			TripInfo:  val,
			StopTimes: formattedTimes,
		})
	}

	return timetable
}
