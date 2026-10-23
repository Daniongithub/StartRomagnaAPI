package scheduler

import (
	"fmt"
	"startromagnaapi/internal/gtfs"
	"startromagnaapi/internal/repository/mezzi"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/sse"

	"github.com/go-co-op/gocron/v2"
)

// internal/scheduler/cron.go — aggiungi in cima al file
import "sync"

var (
	mu      sync.Mutex
	running gocron.Scheduler
)

func Start() error {
	mu.Lock()
	defer mu.Unlock()
	if running != nil {
		return nil
	}
	s, err := InitScheduler()
	if err != nil {
		return err
	}
	running = s
	return nil
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if running == nil {
		return
	}
	running.Shutdown()
	running = nil
}

func InitScheduler() (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	_, err = s.NewJob(gocron.CronJob("0 4 * * *", false), gocron.NewTask(updateStaticTask))

	_, err = s.NewJob(gocron.CronJob("0 0 * * *", false), gocron.NewTask(deleteServiceAlerts))

	_, err = s.NewJob(gocron.CronJob("* * * * *", false), gocron.NewTask(updateServiceAlerts, &sse.SSE_HUB))

	_, err = s.NewJob(gocron.CronJob("*/5 * * * *", false), gocron.NewTask(updateVehiclesStatus))

	_, err = s.NewJob(gocron.CronJob("*/15 * * * * *", true), gocron.NewTask(updateTripUpdates, &sse.SSE_HUB), gocron.WithSingletonMode(gocron.LimitModeReschedule))

	_, err = s.NewJob(gocron.CronJob("*/15 * * * * *", true), gocron.NewTask(updateVehiclePositions, &sse.SSE_HUB), gocron.WithSingletonMode(gocron.LimitModeReschedule))
	
	//_, err = s.NewJob(gocron.CronJob("*/15 * * * * *", true), gocron.NewTask(updateRouteInfos), gocron.WithSingletonMode(gocron.LimitModeReschedule))
	_, err = s.NewJob(gocron.CronJob("0 0 * * *", false), gocron.NewTask(updateRouteInfos), gocron.WithSingletonMode(gocron.LimitModeReschedule))

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

func updateVehiclesStatus() {
	fmt.Println("Task update Vehicles Status")
	mezzi.UpdateVehiclesStatus()
	fmt.Println("Task OK.")
}

func updateRouteInfos() {
	fmt.Println("Task update Route Infos")
	gtfs.UpdateRouteInfos()
	fmt.Println("Task OK.")
}