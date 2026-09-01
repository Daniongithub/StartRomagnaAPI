package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
	"time"
)

func ProcessNextstops(tripId string) model.NextStops {
	nextstops := realtime.GetNextStops(tripId)

	for idx := range nextstops.Stops {
		val := &nextstops.Stops[idx]
		//Adds delay
		val.ArrivalTime = val.ArrivalTime.Add(time.Duration(val.Delay) * time.Second)
		val.DepartureTime = val.DepartureTime.Add(time.Duration(val.Delay) * time.Second)
		//Converts delay into minutes
		val.DelayMin = convertDelay(val.Delay)
		//Formats times
		val.ArrivalTimeStr = val.ArrivalTime.Format("15:04:05")
		val.DepartureTimeStr = val.DepartureTime.Format("15:04:05")
	}

	return nextstops
}

func convertDelay(delay int) int {
    mins := delay / 60
    rest := delay % 60
    if rest >= 30 {
        mins++
    } else if rest <= -30 {
        mins--
    }
    return mins
}
