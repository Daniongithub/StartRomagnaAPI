package model

import "time"

type CorseSopp struct {
	RouteId     string        `json:"route_id"`
	Start       MySQLTime     `json:"start"`
	End         MySQLTime     `json:"end"`
	StartDate   time.Time     `json:"start_date"`
	TripId      string        `json:"trip_id"`
	DirectionId string        `json:"direction_id"`
	FirstStop   CorseSoppStop `json:"first_stop"`
	LastStop    CorseSoppStop `json:"last_stop"`
}

type CorseSoppStop struct {
	StopCode string `db:"stop_code" json:"stop_code"`
	StopName string `db:"stop_name" json:"stop_name"`
}
