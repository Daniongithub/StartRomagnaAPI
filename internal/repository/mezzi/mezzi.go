package mezzi

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/realtime"
)

func GetVehicleInServiceByID(id string) *model.VehicleInService {
	var results []model.VehicleInService
	err := repository.DB_MEZZI.Select(&results, "SELECT matricola, targa, modello, provincia, path_html, photo_path FROM mezzi_start WHERE matricola = ?", id)
	if err != nil {
		fmt.Println("GetVehicleInServiceByID errore db:", err)
	}
	if len(results) == 0 {
		return nil
	}

	return &results[0]
}

func GetMeteInServiceByID(id string) *model.VehicleInService {
	var results []model.VehicleInService
	err := repository.DB_MEZZI.Select(&results, "SELECT matricola, targa, modello, path_html, photo_path FROM mezzi_mete WHERE matricola = ?", id)
	if err != nil {
		fmt.Println("GetVehicleInServiceByID errore db:", err)
	}
	if len(results) == 0 {
		return nil
	}

	return &results[0]
}

func UpdateVehiclesStatus() {
	var buses = realtime.GetVehicles()

	for _, id := range buses {
		_, err := repository.DB_MEZZI.Exec(`UPDATE mezzi_start SET stato = CASE WHEN stato = 'fermo' THEN '' ELSE stato END, last_seen = CURRENT_TIMESTAMP WHERE matricola = ?`, id)
		if err != nil {
			// log errore
			continue
		}
	}
}
