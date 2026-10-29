package realtime

import (
	"fmt"
	"time"

	"startromagnaapi/internal/repository"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/jmoiron/sqlx"
)

func SaveAlerts(feeds map[string]*gtfs.FeedMessage) error {
	return repository.InTransaction(repository.DB_RT, func(tx *sqlx.Tx) error {
		values := make([][]any, 0)
		for basin, feed := range feeds {
			var ids []string
			if err := tx.Select(&ids, "SELECT id FROM service_alerts WHERE basin = ?", basin); err != nil {
				return err
			}
			known := make(map[string]bool, len(ids))
			for _, id := range ids {
				known[id] = true
			}
			for _, entity := range feed.GetEntity() {
				alert := entity.GetAlert()
				if entity.GetIsDeleted() || alert == nil || known[entity.GetId()] {
					continue
				}
				var period *gtfs.TimeRange
				if len(alert.GetActivePeriod()) > 0 {
					period = alert.GetActivePeriod()[0]
				}
				var target *gtfs.EntitySelector
				if len(alert.GetInformedEntity()) > 0 {
					target = alert.GetInformedEntity()[0]
				}
				trip := target.GetTrip()
				var start, end any
				if period.GetStart() != 0 {
					start = time.Unix(int64(period.GetStart()), 0).UTC()
				}
				if period.GetEnd() != 0 {
					end = time.Unix(int64(period.GetEnd()), 0).UTC()
				}
				// The existing schema stores one selector/period for each alert ID.
				var routeID, routeType, tripID, direction, startTime, startDate, stopID any
				if target != nil {
					routeID, routeType, stopID = target.RouteId, target.RouteType, target.StopId
				}
				if trip != nil {
					tripID, direction, startTime, startDate = trip.TripId, trip.DirectionId, trip.StartTime, trip.StartDate
				}
				values = append(values, []any{basin, entity.GetId(), start, end, routeID, routeType, tripID, direction, startTime, startDate, stopID})
				known[entity.GetId()] = true
			}
		}
		return repository.BatchInsert(tx, "service_alerts", []string{"basin", "id", "start", "end", "route_id", "route_type", "trip_id", "direction_id", "start_time", "start_date", "stop_id"}, values)
	})
}

func DeleteAllServiceAlerts() {
	_, err := repository.DB_RT.Exec("DELETE FROM service_alerts")
	if err != nil {
		fmt.Println("DeleteAllServiceAlerts db error:", err)
	}
}
