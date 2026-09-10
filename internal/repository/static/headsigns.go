package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
)

func GetHeadsignsByID(shapeId string) *model.Headsign {
	var results []model.Headsign
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM headsigns WHERE shape_id = ?", shapeId)
	if err != nil {
		fmt.Println("GetHeadsignsByID errore db:", err)
	}
	if len(results) == 0 {
		return nil
	}

	return &results[0]
}

func GetSavedIDs() []model.ShapeID {
	var results []model.ShapeID
	err := repository.DB_STATIC.Select(&results, "SELECT basin, shape_id, still_exists FROM headsigns")
	if err != nil {
		fmt.Println("GetSavedIDs errore db:", err)
	}

	return results
}

func SaveShapeIDs(ids []model.ShapeID) {
	saved := GetSavedIDs()
	savedMap := make(map[string]bool)
	idsMap := make(map[string]bool)
	feedKeys := make(map[string]bool)
	var old []model.ShapeID
	for _, val := range saved {
		savedMap[val.Basin+val.ShapeID] = val.StillExists
	}
	for _, val := range ids {
		idsMap[val.Basin+val.ShapeID] = val.StillExists
	}
	values := make([][]any, 0, len(ids))
	var nowExists []model.ShapeID

	for _, val := range ids {
		active, ok := savedMap[val.Basin+val.ShapeID]
		feedKeys[val.Basin+val.ShapeID] = true
		if !ok {
			values = append(values, []any{
				val.Basin,
				val.ShapeID,
				true,
			})
		}
		if !active {
			nowExists = append(nowExists, model.ShapeID{
				Basin: val.Basin,
				ShapeID: val.ShapeID,
				StillExists: true,
			})
		}
	}

	for _, val := range saved {
		_, ok := feedKeys[val.Basin+val.ShapeID]
		if !ok {
			old = append(old, val)
		}
	}

	//DB insert
	err := repository.BatchInsert(repository.DB_STATIC, "headsigns", []string{"basin", "shape_id", "still_exists"}, values)
	if err != nil {
		fmt.Println("SaveShapeIDs db error:", err)
	}

	//DB update
	for _, val := range nowExists {
		_, err = repository.DB_STATIC.Exec("UPDATE headsigns SET still_exists = 1 WHERE basin = ? AND shape_id = ?", val.Basin, val.ShapeID)
		if err != nil {
			fmt.Println("SaveShapeIDs db error:", err)
		}
	}

	//Still_exists update
	for _, val := range old {
		_, err = repository.DB_STATIC.Exec("UPDATE headsigns SET still_exists = 0 WHERE basin = ? AND shape_id = ?", val.Basin, val.ShapeID)
		if err != nil {
			fmt.Println("SaveShapeIDs db error:", err)
		}
	}
}
