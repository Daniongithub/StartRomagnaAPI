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
	ArrivalTime      MySQLTime `db:"arrival_time" json:"-"`
	ArrivalTimeStr   string    `json:"arrival_time"`
	DepartureTime    MySQLTime `db:"departure_time" json:"-"`
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
	Long         *float32         `db:"longitude" json:"vehicle_long"`
	LastUpdate   time.Time        `db:"timestamp" json:"last_update"`
	NextStop     *StopWDel        `json:"next_stop"`
}

type Arrival struct {
	Basin                string    `db:"basin" json:"basin"`
	Line                 string    `db:"disp_linea" json:"line"`
	Destination          string    `db:"disp_dest" json:"destination"`
	ArrivalTime          MySQLTime `db:"arrival_time" json:"-"`
	ArrivalTimeStr       string    `json:"arrival_time"`
	ScheduleRelationship *string   `db:"schedule_relationship" json:"-"`
	State                string    `json:"state"`
	TripId               string    `db:"trip_id" json:"trip_id"`
	ShapeId              string    `db:"shape_id" json:"shape_id"`
	RouteId              string    `db:"route_id" json:"route_id"`
	OfficialLine         string    `json:"official_line"`
	Vehicle              *string   `db:"vehicle" json:"vehicle"`
	IsGPS                bool      `db:"is_GPS" json:"is_GPS"`
	VehicleConfirmed     bool      `db:"vehicle_confirmed" json:"-"`
	NextStop             *StopWDel `json:"next_stop"`
}

type StopInfo struct {
	Urban    []Line `json:"urban,omitempty"`
	Suburban []Line `json:"suburban,omitempty"`
	Exurban  []Line `json:"exurban,omitempty"`
	Shuttle  []Line `json:"shuttle,omitempty"`
	School   []Line `json:"school,omitempty"`
}

type Line struct {
	Line         string `json:"line"`
	OfficialLine string `json:"official_line"`
	RouteId      string `json:"-"`
}

type LineRow struct {
	ShapeID      string  `db:"shape_id"`
	RouteID      string  `db:"route_id"`
	OfficialLine string  `db:"official_line"`
	DispLine     *string `db:"disp_linea"`
}

type Shape struct {
	Stops  []StopsResult  `json:"stops"`
	Points []ShapesResult `json:"points"`
}

type Announcement struct {
	Id      string `db:"id" json:"-"`
	Message string `db:"message" json:"message"`
}

type RouteInfo struct {
	Basin     string `db:"basin" json:"basin"`
	RouteId   string `db:"route_id" json:"route_id"`
	InfobusId int    `db:"infobus_id" json:"infobus_id"`
	PDFLink   string `db:"pdf_link" json:"pdf_link"`
}
