package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"startromagnaapi/internal/gtfs"
	"startromagnaapi/internal/repository/mezzi"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/sse"

	"github.com/go-co-op/gocron/v2"
)

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
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return nil, err
	}
	s, err := gocron.NewScheduler(gocron.WithLocation(loc))
	if err != nil {
		return nil, err
	}
	jobs := []struct {
		cron    string
		seconds bool
		task    func()
	}{
		{"0 4 * * *", false, func() { run("static", gtfs.UpdateStatic, "") }},
		{"0 0 * * *", false, realtime.DeleteAllServiceAlerts},
		{"* * * * *", false, func() { run("service alerts", gtfs.UpdateAlerts, "service_alerts") }},
		{"*/5 * * * *", false, mezzi.UpdateVehiclesStatus},
		{"*/15 * * * * *", true, func() { run("trip updates", gtfs.UpdateTripUpdates, "trip_updates") }},
		{"*/15 * * * * *", true, func() { run("vehicle positions", gtfs.UpdateVehiclePositions, "vehicle_positions") }},
		{"0 0 * * *", false, func() { run("route infos", gtfs.UpdateRouteInfos, "") }},
		{"0 * * * *", false, func() { run("route infos", clearExpiredAnn, "") }},
	}
	for _, job := range jobs {
		if _, err := s.NewJob(gocron.CronJob(job.cron, job.seconds), gocron.NewTask(job.task), gocron.WithSingletonMode(gocron.LimitModeReschedule)); err != nil {
			s.Shutdown()
			return nil, fmt.Errorf("Schedule %s: %w", job.cron, err)
		}
	}
	s.Start()
	return s, nil
}

func run(name string, update func() error, event string) {
	// start := time.Now()
	if err := update(); err != nil {
		log.Printf("Failed to update %s: %v", name, err)
		return
	}
	if event != "" {
		sse.SSE_HUB.Broadcast(sse.Event{Type: event})
	}
	log.Printf("Updated %s", name)
	// log.Printf("Updated %s in %s", name, time.Since(start))
}

func clearExpiredAnn() error {
	now := time.Now()
	var err error
	announcements := static.GetAnnouncements()
	for _, val := range announcements {
		if val.Expiry.Before(now) {
			err = static.DeleteAnnouncement(val.Id)
		}
	}

	return err
}