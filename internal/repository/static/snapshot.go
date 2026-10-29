package static

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"

	parser "github.com/Leocraft1/gtfsparser-with-reader"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type storedRow struct {
	digest [32]byte
	seen   bool
}

// Update changed rows too: GTFS IDs can remain unchanged while their
// coordinates, times, route names or calendar exceptions change.
// Existing rows are streamed into compact fingerprints; SQL values are built
// only for each write batch, rather than retained for the whole static feed.
func syncTable[T any](tx *sqlx.Tx, table string, columns []string, keys []int, desired []T, values func(T) []any) error {
	identify := func(row []any) string {
		parts := make([]any, len(keys))
		for i, index := range keys {
			parts[i] = row[index]
		}
		data, _ := json.Marshal(parts)
		return string(data)
	}
	fingerprint := func(row []any) ([32]byte, error) {
		data, err := json.Marshal(row)
		return sha256.Sum256(data), err
	}
	rows, err := tx.Queryx("SELECT " + strings.Join(columns, ",") + " FROM " + table)
	if err != nil {
		return err
	}
	old := make(map[string]storedRow)
	for rows.Next() {
		var record T
		if err := rows.StructScan(&record); err != nil {
			rows.Close()
			return err
		}
		row := values(record)
		digest, err := fingerprint(row)
		if err != nil {
			rows.Close()
			return err
		}
		old[identify(row)] = storedRow{digest: digest}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	insert := make([]int, 0)
	remove := make([]string, 0)
	for i, record := range desired {
		row := values(record)
		key := identify(row)
		previous, exists := old[key]
		if previous.seen {
			return fmt.Errorf("Duplicate key in %s: %s", table, key)
		}
		digest, err := fingerprint(row)
		if err != nil {
			return err
		}
		if !exists || previous.digest != digest {
			if exists {
				remove = append(remove, key)
			}
			insert = append(insert, i)
		}
		previous.seen = true
		old[key] = previous
	}
	for key, previous := range old {
		if !previous.seen {
			remove = append(remove, key)
		}
	}
	for start := 0; start < len(remove); start += 500 {
		conditions := make(squirrel.Or, 0, 500)
		for _, key := range remove[start:min(start+500, len(remove))] {
			var parts []any
			if err := json.Unmarshal([]byte(key), &parts); err != nil {
				return err
			}
			where := squirrel.Eq{}
			for i, index := range keys {
				where[columns[index]] = parts[i]
			}
			conditions = append(conditions, where)
		}
		query, args, err := squirrel.Delete(table).Where(conditions).ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}
	for start := 0; start < len(insert); start += 5000 {
		batch := make([][]any, 0, min(5000, len(insert)-start))
		for _, index := range insert[start:min(start+5000, len(insert))] {
			batch = append(batch, values(desired[index]))
		}
		if err := repository.BatchInsert(tx, table, columns, batch); err != nil {
			return err
		}
	}
	return nil
}

func SaveSnapshot(feeds map[string]*parser.Feed) error {
	for _, basin := range []string{"RA", "FC", "RN"} {
		if feeds[basin] == nil {
			return fmt.Errorf("Missing static feed %s", basin)
		}
	}
	return repository.InTransaction(repository.DB_STATIC, func(tx *sqlx.Tx) error {
		var trips []model.TripsResult
		var routes []model.RoutesResult
		var shapes []model.ShapesResult
		var stops []model.StopsResult
		var times []model.StopTimesResult
		var dates []model.CalendarDatesResult
		for _, basin := range []string{"RA", "FC", "RN"} {
			feed := feeds[basin]
			for _, trip := range feed.Trips {
				row := model.ToDomainTrip(trip)
				row.Basin = basin
				trips = append(trips, row)
				for _, stop := range trip.StopTimes {
					arrival, departure := stop.Arrival_time(), stop.Departure_time()
					times = append(times, model.ToDomainStopTimes(basin, trip.Id,
						time.Date(0, 1, 1, int(arrival.Hour), int(arrival.Minute), int(arrival.Second), 0, time.UTC),
						time.Date(0, 1, 1, int(departure.Hour), int(departure.Minute), int(departure.Second), 0, time.UTC), stop.Stop().Id, stop.Sequence()))
				}
			}
			for _, route := range feed.Routes {
				row := model.ToDomainRoutes(route)
				row.Basin = basin
				routes = append(routes, row)
			}
			for _, shape := range feed.Shapes {
				for _, point := range shape.Points {
					shapes = append(shapes, model.ToDomainShapes(basin, shape.Id, point.Lat, point.Lon, int(point.Sequence)))
				}
			}
			for _, stop := range feed.Stops {
				row := model.ToDomainStops(stop)
				row.Basin = basin
				stops = append(stops, row)
			}
			for _, service := range feed.Services {
				// These feeds provide explicit dates in calendar_dates.txt only.
				for date, added := range service.Exceptions() {
					dates = append(dates, model.ToDomainException(service, basin, date, added))
				}
			}
		}
		// Preserve manually flagged dummy stops while marking new synthetic stops.
		var dummyStops []model.StopsResult
		if err := tx.Select(&dummyStops, "SELECT basin, stop_id FROM stops WHERE is_dummy = 1"); err != nil {
			return err
		}
		dummy := make(map[[2]string]bool, len(dummyStops))
		for _, stop := range dummyStops {
			dummy[[2]string{stop.Basin, stop.Stop_id}] = true
		}
		for i := range stops {
			stop := &stops[i]
			name := strings.ToLower(stop.Stop_name)
			stop.Is_dummy = dummy[[2]string{stop.Basin, stop.Stop_id}] || strings.Contains(name, "semaforo") || strings.Contains(name, "fi1") || strings.Contains(name, "fittizio") || strings.Contains(name, "fittizia")
		}
		operations := []func() error{
			func() error {
				return syncTable(tx, "trips", []string{"basin", "route_id", "service_id", "trip_id", "trip_headsign", "direction_id", "shape_id"}, []int{0, 3}, trips, func(v model.TripsResult) []any {
					return []any{v.Basin, v.Route_id, v.Service_id, v.Trip_id, v.Trip_headsign, v.Direction_id, v.Shape_id}
				})
			},
			func() error {
				return syncTable(tx, "routes", []string{"basin", "route_id", "agency_id", "route_short_name", "route_long_name", "route_type"}, []int{0, 1}, routes, func(v model.RoutesResult) []any {
					return []any{v.Basin, v.Route_id, v.Agency_id, v.Route_short_name, v.Route_long_name, v.Route_type}
				})
			},
			func() error {
				return syncTable(tx, "shapes", []string{"basin", "shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence"}, []int{0, 1, 4}, shapes, func(v model.ShapesResult) []any {
					return []any{v.Basin, v.Shape_id, v.Shape_pt_lat, v.Shape_pt_lon, v.Shape_pt_sequence}
				})
			},
			func() error {
				return syncTable(tx, "stops", []string{"basin", "stop_id", "stop_code", "stop_name", "stop_lat", "stop_lon", "is_dummy"}, []int{0, 1}, stops, func(v model.StopsResult) []any {
					return []any{v.Basin, v.Stop_id, v.Stop_code, v.Stop_name, v.Stop_lat, v.Stop_lon, v.Is_dummy}
				})
			},
			func() error {
				return syncTable(tx, "stop_times", []string{"basin", "trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence"}, []int{0, 1, 5}, times, func(v model.StopTimesResult) []any {
					return []any{v.Basin, v.Trip_id, v.Arrival_time.SQLString(), v.Departure_time.SQLString(), v.Stop_id, v.Stop_sequence}
				})
			},
			func() error {
				return syncTable(tx, "calendar_dates", []string{"basin", "service_id", "date", "exception_type"}, []int{0, 1, 2}, dates, func(v model.CalendarDatesResult) []any {
					return []any{v.Basin, v.Service_id, v.Date.Format("2006-01-02"), v.Exception_type}
				})
			},
		}
		for _, update := range operations {
			if err := update(); err != nil {
				return err
			}
		}
		return syncHeadsigns(tx)
	})
}

func syncHeadsigns(tx *sqlx.Tx) error {
	if _, err := tx.Exec(`UPDATE headsigns h SET still_exists = EXISTS (SELECT 1 FROM shapes s WHERE s.basin = h.basin AND s.shape_id = h.shape_id)`); err != nil {
		return err
	}
	_, err := tx.Exec(`INSERT INTO headsigns (basin, shape_id, still_exists)
	 SELECT DISTINCT s.basin, s.shape_id, 1 FROM shapes s
	 WHERE NOT EXISTS (SELECT 1 FROM headsigns h WHERE h.basin = s.basin AND h.shape_id = s.shape_id)`)
	return err
}
