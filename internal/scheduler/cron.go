package scheduler

import (
	"fmt"
	"startromagnaapi/internal/gtfs"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/sse"

	"github.com/go-co-op/gocron/v2"
)

func InitScheduler(hub *sse.Hub) (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	_, err = s.NewJob(gocron.CronJob("0 4 * * *", false), gocron.NewTask(updateStaticTask))

	_, err = s.NewJob(gocron.CronJob("0 0 * * *", false), gocron.NewTask(deleteServiceAlerts))

	_, err = s.NewJob(gocron.CronJob("* * * * *", false), gocron.NewTask(updateServiceAlerts, hub))

	_, err = s.NewJob(gocron.CronJob("*/30 * * * * *", true), gocron.NewTask(updateTripUpdates, hub), gocron.WithSingletonMode(gocron.LimitModeReschedule))

	_, err = s.NewJob(gocron.CronJob("*/30 * * * * *", true), gocron.NewTask(updateVehiclePositions, hub), gocron.WithSingletonMode(gocron.LimitModeReschedule))

	if err != nil {
		return nil, err
	}

	s.Start()
	return s, nil
}

func updateStaticTask() {
	fmt.Println("Task delle 04:00, aggiorno dati GTFS statici...")
	gtfs.UpdateStatic()
	fmt.Println("Task OK.")
}

func updateServiceAlerts(hub *sse.Hub) {
	fmt.Println("Task update Service Alerts")
	gtfs.UpdateAlerts()
	hub.Broadcast(sse.Event{Type: "service_alerts"})
	fmt.Println("Task OK.")
}

func deleteServiceAlerts() {
	fmt.Println("Task delete Service Alerts")
	realtime.DeleteAllServiceAlerts()
	fmt.Println("Task OK.")
}

func updateTripUpdates(hub *sse.Hub) {
	fmt.Println("Task update Trip Updates")
	gtfs.UpdateTripUpdates()
	hub.Broadcast(sse.Event{Type: "trip_updates"})
	fmt.Println("Task OK.")
}

func updateVehiclePositions(hub *sse.Hub) {
	fmt.Println("Task update Vehicle Positions")
	gtfs.UpdateVehiclePositions()
	hub.Broadcast(sse.Event{Type: "vehicle_positions"})
	fmt.Println("Task OK.")
}
