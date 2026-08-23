package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"
)

func GetStopTimes() []model.StopTimesResult {
	var results []model.StopTimesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM stop_times")
	if err != nil {
		fmt.Println("GetStopTimes errore db:", err)
	}

	return results
}