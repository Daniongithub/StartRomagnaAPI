package service

import (
	"sort"
	"strings"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
)

func ProcessTimetable(basin, routeId, directionId string) model.Timetable {
	trips := static.GetTodayTripInfoFromRouteID(basin, routeId, directionId)
	timetable := model.Timetable{
		Stops: make([]model.StopTimeWName, 0),
		Trips: make([]model.Trip, 0, len(trips)),
	}
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
		timetable.Trips = append(timetable.Trips, model.Trip{
			TripInfo:  val,
			StopTimes: formattedTimes,
		})
		// Retain every passage, including repeated stops within or across trips.
		timetable.Stops = append(timetable.Stops, formattedTimes...)
	}

	sort.SliceStable(timetable.Stops, func(i, j int) bool {
		a, b := &timetable.Stops[i], &timetable.Stops[j]
		if a.StopSequence != b.StopSequence {
			return a.StopSequence < b.StopSequence
		}
		arrivalA, arrivalB := a.ArrivalTime.ServiceDuration(), b.ArrivalTime.ServiceDuration()
		if arrivalA != arrivalB {
			return arrivalA < arrivalB
		}
		return a.TripTd < b.TripTd
	})
	return timetable
}
