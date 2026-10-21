package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"strings"
)

func ProcessTimetable(basin, routeId, directionId string) []model.Trip {
	trips := static.GetTodayTripInfoFromRouteID(basin, routeId, directionId)
	var timetable []model.Trip
	for _, val := range trips {
		val.OfficialLine = static.GetRouteNamefromID(val.Basin, routeId)
		headsign := static.GetHeadsignsByID(val.Basin, val.ShapeId)
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
		//Formats arrival time
		formattedTimes := static.GetStopTimesDetailByTrip(basin, val.TripId)
		for idx := range formattedTimes {
			val := &formattedTimes[idx]
			val.ArrivalTimeStr = val.ArrivalTime.Format("15:04")
			val.DepartureTimeStr = val.DepartureTime.Format("15:04")
		}
		timetable = append(timetable, model.Trip{
			TripInfo: val,
			StopTimes: formattedTimes,
		})
	}

	return timetable
}