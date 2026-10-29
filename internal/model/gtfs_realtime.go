package model

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type MySQLTime struct {
	time.Time
}

func (mt *MySQLTime) Scan(value interface{}) error {
	if value == nil {
		mt.Time = time.Time{}
		return nil
	}

	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	case time.Time:
		mt.Time = v
		return nil
	default:
		return fmt.Errorf("Unsupported Scan type for MySQLTime: %T", value)
	}

	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("Invalid MySQL TIME %q", raw)
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 || hours > 838 {
		return fmt.Errorf("Invalid GTFS hours %q", raw)
	}
	rest, err := time.Parse("04:05", parts[1])
	if err != nil {
		return fmt.Errorf("Invalid MySQL TIME %q: %w", raw, err)
	}
	mt.Time = time.Date(0, 1, 1, hours, rest.Minute(), rest.Second(), rest.Nanosecond(), time.UTC)
	return nil
}

func (mt MySQLTime) Value() (driver.Value, error) {
	return mt.SQLString(), nil
}

// SQLString preserves GTFS service hours above 24:00:00.
func (mt MySQLTime) SQLString() string {
	hours := mt.Hour()
	if mt.Year() == 0 {
		hours += (mt.YearDay() - 1) * 24
	}
	return fmt.Sprintf("%02d:%02d:%02d", hours, mt.Minute(), mt.Second())
}

func (mt MySQLTime) ServiceDuration() time.Duration {
	if mt.Year() == 0 {
		return mt.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))
	}
	return time.Duration(mt.Hour())*time.Hour + time.Duration(mt.Minute())*time.Minute + time.Duration(mt.Second())*time.Second
}
