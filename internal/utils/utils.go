package utils

import (
	"startromagnaapi/internal/model"
	"time"
)

func ConvertDelay(delay int) int {
    mins := delay / 60
    rest := delay % 60
    if rest >= 30 {
        mins++
    } else if rest <= -30 {
        mins--
    }
    return mins
}

func FixStopWDel(stops []model.StopWDel) {
	for idx := range stops {
		val := &stops[idx]
		//Adds delay
		val.ArrivalTime.Time = val.ArrivalTime.Add(time.Duration(val.Delay) * time.Second)
		val.DepartureTime.Time = val.DepartureTime.Add(time.Duration(val.Delay) * time.Second)
		//Converts delay into minutes
		val.DelayMin = ConvertDelay(val.Delay)
		//Formats times
		val.ArrivalTimeStr = val.ArrivalTime.Format("15:04:05")
		val.DepartureTimeStr = val.DepartureTime.Format("15:04:05")
	}
}