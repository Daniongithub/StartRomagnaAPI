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

	for _, val := range buses {
		_, err := repository.DB_MEZZI.Exec(`UPDATE mezzi_start SET stato = CASE WHEN stato IN ('fermo', 'sconosciuto') THEN '' ELSE stato END, last_seen = CURRENT_TIMESTAMP, provincia = ? WHERE matricola = ?`, val.Basin, val.Number)
		if err != nil {
			fmt.Println("UpdateVehiclesStatus errore db:", err)
			continue
		}
	}

	// Imposta "sconosciuto" ai mezzi non visti da almeno 5 giorni
	_, err := repository.DB_MEZZI.Exec(`UPDATE mezzi_start SET stato = 'sconosciuto' WHERE last_seen IS NOT NULL AND last_seen < CURRENT_TIMESTAMP - INTERVAL 5 DAY`)

	if err != nil {
		fmt.Println("UpdateVehiclesStatus errore db:", err)
	}
}
