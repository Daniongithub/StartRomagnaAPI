package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
)

func GetRouteInfos() []model.RouteInfo {
	var results []model.RouteInfo
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM route_infos")
	if err != nil {
		fmt.Println("GetRouteInfos error:", err)
	}

	return results
}

func SaveRouteInfos(infos []model.RouteInfo) {
	for _, val := range infos {
		_, err := repository.DB_STATIC.Exec("INSERT INTO route_infos VALUES(?, ?, ?, ?)", val.Basin, val.RouteId, val.InfobusId, val.PDFLink)
		if err != nil {
			fmt.Println("SaveRouteInfos error:", err, val.RouteId)
		}
	}
}

func ClearRouteInfos() {
	_, err := repository.DB_STATIC.Exec("DELETE FROM route_infos")
	if err != nil {
		fmt.Println("GetRouteInfos error:", err)
	}
}
