package service

import (
	"errors"
	"sort"
	"time"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/utils"
)

var NoArrivals = errors.New("No arrivals")

func ProcessArrivals(stopCode string) ([]model.Arrival, error) {
	arrivals := static.GetArrivals(stopCode)
	if len(arrivals) == 0 {
		return nil, NoArrivals
	}
	now := utils.RomeNow()
	var results []model.Arrival
	enrich := newBusEnricher()
	for idx := range arrivals {
		val := &arrivals[idx]
		if val.VehicleConfirmed &&
			(val.ScheduleRelationship == nil || *val.ScheduleRelationship != "CANCELED") {
			val.NextStop = realtime.GetFirstStop(val.TripId)
		}
		if !finalizeArrival(val, now) {
			continue
		}
		val.OfficialLine, val.Line, val.Destination = enrich.lineInfo(val.Basin, val.RouteId, val.ShapeId, val.TripId)
		results = append(results, *val)
	}
	if len(results) == 0 {
		return nil, NoArrivals
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].ArrivalTime.Before(results[j].ArrivalTime.Time) })
	return results, nil
}

// Resolve the service day before applying the next stop's rounded delay, so extended GTFS hours
// and delays that cross midnight keep the correct calendar date.
func finalizeArrival(val *model.Arrival, now time.Time) bool {
	date := val.ServiceDate
	val.ArrivalTime.Time = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(val.ArrivalTime.ServiceDuration())
	val.State = "planned"
	canceled := val.ScheduleRelationship != nil && *val.ScheduleRelationship == "CANCELED"
	if canceled {
		val.State, val.Vehicle, val.NextStop, val.IsGPS = "canceled", nil, nil, false
	} else if val.VehicleConfirmed && val.NextStop != nil {
		val.State = "realtime"
		val.ArrivalTime.Time = val.ArrivalTime.Add(time.Duration(val.NextStop.DelayMin) * time.Minute)
	}
	wallNow := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
	val.ArrivalTimeStr = val.ArrivalTime.Format("15:04")
	return !val.ArrivalTime.Before(wallNow)
}
