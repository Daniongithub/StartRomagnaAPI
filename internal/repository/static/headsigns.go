package static

import (
	"fmt"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
)

func GetHeadsignsByID(basin, shapeId string) *model.Headsign {
	var results []model.Headsign
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM headsigns WHERE shape_id = ? AND basin = ?", shapeId, basin)
	if err != nil {
		fmt.Println("GetHeadsignsByID database error:", err)
	}
	if len(results) == 0 {
		return nil
	}

	return &results[0]
}
