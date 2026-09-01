package model

import "time"

type CorseSopp struct {
	RouteId     string        `json:"route_id"`
	Start       string        `json:"start"`
	End         string        `json:"end"`
	StartDate   string        `json:"start_date"`
	TripId      string        `json:"trip_id"`
	DirectionId int           `json:"direction_id"`
	FirstStop   CorseSoppStop `json:"first_stop"`
	LastStop    CorseSoppStop `json:"last_stop"`
	Vehicle     *string       `json:"vehicle"`
}

type CorseSoppStop struct {
	StopCode string `db:"stop_code" json:"stop_code"`
	StopName string `db:"stop_name" json:"stop_name"`
}

type NextStops struct {
	Stops []StopWDel `json:"stops"`
}

type StopWDel struct {
	Basin            string    `db:"basin" json:"basin"`
	StopId           string    `db:"stop_id" json:"stop_id"`
	StopCode         string    `db:"stop_code" json:"stop_code"`
	StopName         string    `db:"stop_name" json:"stop_name"`
	StopLat          float32   `db:"stop_lat" json:"stop_lat"`
	StopLon          float32   `db:"stop_lon" json:"stop_lon"`
	ArrivalTime      time.Time `db:"arrival_time" json:"-"`
	ArrivalTimeStr   string    `json:"arrival_time"`
	DepartureTime    time.Time `db:"departure_time" json:"-"`
	DepartureTimeStr string    `json:"departure_time"`
	Delay            int       `db:"delay" json:"-"`
	DelayMin         int       `json:"delay"`
}
