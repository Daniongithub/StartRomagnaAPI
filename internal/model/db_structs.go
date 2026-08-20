package model

type TripsResult struct {
	Route_id string `db:"route_id" json:"route_id"`
	Service_id int `db:"service_id" json:"service_id"`
	Trip_id string `db:"trip_id" json:"trip_id"`
	Trip_headsign *string `db:"trip_headsign" json:"trip_headsign"`
	Direction_id int `db:"direction_id" json:"direction_id"`
	Shape_id int `db:"shape_id" json:"shape_id"`
}