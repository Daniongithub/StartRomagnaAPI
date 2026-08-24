package repository

import (
	"StartRomagnaAPI/internal/model"
	"fmt"
)

func GetShapes() []model.ShapesResult {
	var results []model.ShapesResult
	err := DB_CONTENT.Select(&results, "SELECT * FROM shapes")
	if err != nil {
		fmt.Println("GetShapes errore db:", err)
	}

	return results
}
