package model

import "time"

type ServiceAlertsResult struct {
	Basin      string    `db:"basin" json:"basin"`
	Id         string    `db:"id" json:"id"`
	Start      time.Time `db:"start" json:"start"`
	End        time.Time `db:"end" json:"end"`
	Route_id   string    `db:"route_id" json:"route_id"`
	Route_type string    `db:"route_type" json:"route_type"`
	Start_time time.Time `db:"start_time" json:"start_time"`
	Start_date time.Time `db:"start_date" json:"start_date"`
	Stop_id    string    `db:"stop_id" json:"stop_id"`
}
