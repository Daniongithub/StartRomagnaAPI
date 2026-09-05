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
		val.NextStop = realtime.GetFirstStop(val.TripId)
		//Confirms trip state
		if val.NextStop != nil {
			val.State = "realtime"
			//Adds delay to arrival time
			val.ArrivalTime.Time = val.ArrivalTime.Add(time.Duration(val.NextStop.DelayMin) * time.Minute)
		} else {
			val.State = "planned"
		}
		val.ArrivalTimeStr = val.ArrivalTime.Format("15:04")
	}

	return arrivals
}
