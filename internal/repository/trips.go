package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"
)

func GetTrips() []model.TripsResult {
	var results []model.TripsResult
	err := DB_CONTENT.Select(&results,"SELECT * FROM trips")
	if err != nil {
		fmt.Println("GetTrips error:", err)
	}

	return results
}