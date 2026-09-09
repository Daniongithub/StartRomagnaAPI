package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"strings"
	"time"
)

func ProcessArrivals(stopCode string) []model.Arrival {
	arrivals := static.GetArrivals(stopCode)
	now := time.Now().Add(1 * time.Minute)
	today := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
	var results []model.Arrival

	//Fixes and adds arrivals information
	for idx := range arrivals {
		val := &arrivals[idx]
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
		//If the vehicle didn't reverse himself (but it's communicating the position) remove next stop and delay indication
		if val.VehicleConfirmed {
			val.NextStop = realtime.GetFirstStop(val.TripId)
		}
		//Confirms trip state
		if val.NextStop != nil {
			val.State = "realtime"
			//Adds delay to arrival time
			val.ArrivalTime.Time = val.ArrivalTime.Add(time.Duration(val.NextStop.DelayMin) * time.Minute)
		} else {
			val.State = "planned"
		}
		//Show canceled status
		if val.ScheduleRelationship != nil && *val.ScheduleRelationship == "CANCELED" {
			val.State = "canceled"
			val.Vehicle = nil
			val.NextStop = nil
		}

		// Ricostruisce la data corretta partendo dall'orario (ignorando l'anno fittizio 0000-01-01)
		val.ArrivalTime.Time = time.Date(
			today.Year(), today.Month(), today.Day(),
			val.ArrivalTime.Hour(), val.ArrivalTime.Minute(), val.ArrivalTime.Second(), 0,
			time.UTC,
		)

		//Removes trip if bus has already went through the stop (realtime arrival < now)
		if val.ArrivalTime.Before(today) {
			continue
		}

		val.ArrivalTimeStr = val.ArrivalTime.Format("15:04")
		results = append(results, *val)
	}

	return results
}
