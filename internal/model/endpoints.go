package model

import (
	"time"
)

type NextStops struct {
	Stops []StopWDel `json:"stops"`
}

type StopWDel struct {
	Basin            string    `db:"basin" json:"-"`
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
	Error            string    `json:"error,omitempty"`
}

type BusInService struct {
	Basin        string           `db:"basin" json:"basin"`
	Line         string           `db:"disp_linea" json:"line"`
	Destination  string           `db:"disp_dest" json:"destination"`
	TripId       string           `db:"trip_id" json:"trip_id"`
	ShapeId      string           `db:"shape_id" json:"shape_id"`
	RouteId      string           `db:"route_id" json:"route_id"`
	OfficialLine string           `json:"official_line"`
	Vehicle      string           `db:"vehicle" json:"-"`
	VehicleInfo  VehicleInService `json:"vehicle_info"`
	Lat          *float32         `db:"lat" json:"vehicle_lat"`
	Long         *float32         `db:"long" json:"vehicle_long"`
	LastUpdate   time.Time        `db:"timestamp" json:"last_update"`
	NextStop     StopWDel         `json:"next_stop"`
}

type VehiclePosition struct {
	Basin        string           `db:"basin" json:"basin"`
	Line         string           `db:"disp_linea" json:"line"`
	Destination  string           `db:"disp_dest" json:"destination"`
	TripId       string           `db:"trip_id" json:"trip_id"`
	ShapeId      string           `db:"shape_id" json:"shape_id"`
	RouteId      string           `db:"route_id" json:"route_id"`
	OfficialLine string           `json:"official_line"`
	Vehicle      string           `db:"vehicle" json:"-"`
	VehicleInfo  VehicleInService `json:"vehicle_info"`
	Lat          float32          `db:"lat" json:"lat"`
	Long         float32          `db:"long" json:"long"`
	LastUpdate   time.Time        `db:"timestamp" json:"last_update"`
}
