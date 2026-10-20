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
	values := make([][]any, 0, len(infos))

	for _, val := range infos {
		values = append(values, []any{
			val.Basin,
			val.RouteId,
			val.InfobusId,
			val.PDFLink,
		})
	}

	err := repository.BatchInsert(repository.DB_STATIC, "routes", []string{"basin", "route_id", "infobus_id", "pdf_link"}, values)
	if err != nil {
		fmt.Println("SaveRouteInfos db error:", err)
	}
}

func ClearRouteInfos() {
	_, err := repository.DB_STATIC.Exec("DELETE FROM route_infos")
	if err != nil {
		fmt.Println("GetRouteInfos error:", err)
	}
}
