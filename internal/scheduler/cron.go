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
