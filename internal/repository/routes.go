package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"
)

func GetRoutes() []model.RoutesResult {
	var results []model.RoutesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM routes")
	if err != nil {
		fmt.Println("GetRoutes errore db:", err)
	}

	return results
}