package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type ServiceAlertsResult struct {
	Basin        string    `db:"basin" json:"basin"`
	Id           string    `db:"id" json:"id"`
	Start        MySQLTime `db:"start" json:"start"`
	End          MySQLTime `db:"end" json:"end"`
	Route_id     string    `db:"route_id" json:"route_id"`
	Route_type   string    `db:"route_type" json:"route_type"`
	Trip_id      string    `db:"trip_id" json:"trip_id"`
	Direction_id int       `db:"direction_id" json:"direction_id"`
	Start_time   MySQLTime `db:"start_time" json:"start_time"`
	Start_date   time.Time `db:"start_date" json:"start_date"`
	Stop_id      string    `db:"stop_id" json:"stop_id"`
}

type TripUpdatesResult struct {
	Basin                 string    `db:"basin" json:"basin"`
	Id                    string    `db:"id" json:"id"`
	Trip_id               string    `db:"trip_id" json:"trip_id"`
	Route_id              string    `db:"route_id" json:"route_id"`
	Start                 MySQLTime `db:"start" json:"start"`
	End                   MySQLTime `db:"end" json:"end"`
	Direction_id          int       `db:"direction_id" json:"direction_id"`
	Start_time            MySQLTime `db:"start_time" json:"start_time"`
	Start_date            time.Time `db:"start_date" json:"start_date"`
	Schedule_relationship *string   `db:"schedule_relationship" json:"schedule_relationship"`
	Vehicle               int       `db:"vehicle" json:"vehicle"`
	Timestamp             time.Time `db:"timestamp" json:"timestamp"`
	Stop_time_updates     []StopTime
}

type StopTime struct {
	Trip_id       string `db:"trip_id" json:"trip_id"`
	Stop_sequence int    `db:"stop_sequence" json:"stop_sequence"`
	Delay         int    `db:"delay" json:"delay"`
}

type MySQLTime struct {
	time.Time
}

func (mt *MySQLTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("Unsupported Scan type for MySQLTime: %T", value)
	}

	parsed, err := time.Parse("15:04:05", string(raw))
	if err != nil {
		// prova anche il formato con microsecondi
		parsed, err = time.Parse("15:04:05.999999", string(raw))
		if err != nil {
			return fmt.Errorf("Cannot parse time %q: %w", raw, err)
		}
	}

	mt.Time = parsed
	return nil
}

func (mt MySQLTime) Value() (driver.Value, error) {
	return mt.Time.Format("15:04:05"), nil
}
