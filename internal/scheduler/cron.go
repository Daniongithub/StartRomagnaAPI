package scheduler

import (
	"StartRomagnaAPI/internal/gtfs"
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

func InitScheduler() (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	_, err = s.NewJob(gocron.CronJob("0 4 * * *", false), gocron.NewTask(updateStaticTask))

	_, err = s.NewJob(gocron.CronJob("* * * * *", false), gocron.NewTask(updateServiceAlerts))

	_, err = s.NewJob(gocron.CronJob("*/20 * * * * *", true), gocron.NewTask(updateTripUpdates), gocron.WithSingletonMode(gocron.LimitModeReschedule))

	_, err = s.NewJob(gocron.CronJob("*/20 * * * * *", true), gocron.NewTask(updateVehiclePositions))

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

func updateServiceAlerts() {
	fmt.Println("Task update Service Alerts")
	gtfs.UpdateAlerts()
	fmt.Println("Task OK.")
}

func updateTripUpdates() {
	fmt.Println("Task update Trip Updates")
	gtfs.UpdateTripUpdates()
	fmt.Println("Task OK.")
}

func updateVehiclePositions() {
	fmt.Println("Task update ")
	gtfs.UpdateVehiclePositions()
	fmt.Println("Task OK.")
}
