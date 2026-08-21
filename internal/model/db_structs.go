package model

import (
	"time"

	"github.com/Leocraft1/gtfsparser-with-reader/gtfs"
)

type TripsResult struct {
	Basin         string  `db:"basin" json:"basin"`
	Route_id      string  `db:"route_id" json:"route_id"`
	Service_id    string  `db:"service_id" json:"service_id"`
	Trip_id       string  `db:"trip_id" json:"trip_id"`
	Trip_headsign *string `db:"trip_headsign" json:"trip_headsign"`
	Direction_id  int     `db:"direction_id" json:"direction_id"`
	Shape_id      string  `db:"shape_id" json:"shape_id"`
}

func ToDomainTrip(r *gtfs.Trip) TripsResult {
	return TripsResult{
		Route_id:      r.Route.Id,
		Service_id:    r.Service.Id(),
		Trip_id:       r.Id,
		Trip_headsign: r.Headsign,
		Direction_id:  int(r.Direction_id),
		Shape_id:      r.Shape.Id,
	}
}

type CalendarDatesResult struct {
	Basin          string    `db:"basin" json:"basin"`
	Service_id     string    `db:"service_id" json:"service_id"`
	Date           time.Time `db:"date" json:"date"`
	Exception_type string    `db:"exception_type" json:"exception_type"`
}

func ToDomainException(r *gtfs.Service, basin string, date gtfs.Date, added bool) CalendarDatesResult {
	return CalendarDatesResult{
		Basin:			basin,
		Service_id:     r.Id(),
		Date:           date.GetTime(),
		Exception_type: convertExceptionType(added),
	}
}

func convertExceptionType(ex bool) string {
	if ex {
		return "1"
	}
	return "2"
}