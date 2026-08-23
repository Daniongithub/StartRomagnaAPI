package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"
)

func GetStops() []model.StopsResult {
	var results []model.StopsResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM stops")
	if err != nil {
		fmt.Println("GetStops errore db:", err)
	}

	return results
}