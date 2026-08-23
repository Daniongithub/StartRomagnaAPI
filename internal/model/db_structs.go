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
		Basin:          basin,
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

type RoutesResult struct {
	Basin            string `db:"basin" json:"basin"`
	Route_id         string `db:"route_id" json:"route_id"`
	Agency_id        string `db:"agency_id" json:"agency_id"`
	Route_short_name string `db:"route_short_name" json:"route_short_name"`
	Route_long_name  string `db:"route_long_name" json:"route_long_name"`
	Route_type       int    `db:"route_type" json:"route_type"`
}

type ShapesResult struct {
	Basin             string `db:"basin" json:"basin"`
	Shape_id          int    `db:"shape_id" json:"shape_id"`
	Shape_pt_lat      string `db:"shape_pt_lat" json:"shape_pt_lat"`
	Shape_pt_lon      string `db:"shape_pt_lon" json:"shape_pt_lon"`
	Shape_pt_sequence string `db:"shape_pt_sequence" json:"shape_pt_sequence"`
}

type StopTimesResult struct {
	Basin          string    `db:"basin" json:"basin"`
	Trip_id        string    `db:"trip_id" json:"trip_id"`
	Arrival_time   time.Time `db:"arrival_time" json:"arrival_time"`
	Departure_time time.Time `db:"departure_time" json:"departure_time"`
	Stop_id        int       `db:"stop_id" json:"stop_id"`
	Stop_sequence  int       `db:"stop_sequence" json:"stop_sequence"`
}

type StopsResult struct {
	Basin     string  `db:"basin" json:"basin"`
	Stop_id   int     `db:"stop_id" json:"stop_id"`
	Stop_code int     `db:"stop_code" json:"stop_code"`
	Stop_name string  `db:"stop_name" json:"stop_name"`
	Stop_lat  float32 `db:"stop_lat" json:"stop_lat"`
	Stop_lon  float32 `db:"stop_lon" json:"stop_lon"`
}
