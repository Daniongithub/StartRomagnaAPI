package mezzi

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
)

func GetVehicleInServiceByID(id string) *model.VehicleInService {
	var results []model.VehicleInService
	err := repository.DB_MEZZI.Select(&results, "SELECT matricola, targa, modello, provincia, photo_path FROM mezzi_start WHERE matricola = ?", id)
	if err != nil {
		fmt.Println("GetVehicleInServiceByID errore db:", err)
	}
	if len(results) == 0 {
		return nil
	}

	return &results[0]
}