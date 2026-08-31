package model

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

