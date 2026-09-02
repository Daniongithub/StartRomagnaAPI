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
	err := repository.DB_STATIC.Select(&results, "SELECT basin, shape_id FROM headsigns")
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
		savedMap[val.Basin+val.ShapeID] = true
	}
	for _, val := range ids {
		idsMap[val.Basin+val.ShapeID] = true
	}
	values := make([][]any, 0, len(ids))

	for _, val := range ids {
		_, ok := savedMap[val.Basin+val.ShapeID]
		feedKeys[val.Basin+val.ShapeID] = true
		if !ok {
			values = append(values, []any{
				val.Basin,
				val.ShapeID,
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
	err := repository.BatchInsert(repository.DB_STATIC, "headsigns", []string{"basin", "shape_id"}, values)
	if err != nil {
		fmt.Println("SaveShapeIDs db error:", err)
	}

	//Database delete
	for _, val := range old {
		_, err = repository.DB_STATIC.Exec("DELETE FROM headsigns WHERE basin = ? AND shape_id = ?", val.Basin, val.ShapeID)
		if err != nil {
			fmt.Println("SaveShapeIDs db error:", err)
		}
	}
}
