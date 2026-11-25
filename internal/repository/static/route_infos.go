package static

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/notify"
	"startromagnaapi/internal/repository"

	"github.com/jmoiron/sqlx"
)

func GetRouteInfos() []model.RouteInfo {
	var results []model.RouteInfo
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM route_infos")
	if err != nil {
		notify.ReportError("Failed to read route information", err)
	}

	return results
}

func GetInfoBusId(routeId string) []string {
	var results []string
	err := repository.DB_STATIC.Select(&results, "SELECT infobus_id FROM route_infos WHERE route_id = ?", routeId)
	if err != nil {
		notify.ReportError("Failed to read a route's InfoBus ID", err)
	}

	return results
}

func SaveRouteInfos(tx *sqlx.Tx, infos []model.RouteInfo) error {
	values := make([][]any, 0, len(infos))

	for _, val := range infos {
		values = append(values, []any{
			val.Basin,
			val.RouteId,
			val.StartText,
			val.InfobusId,
			val.PDFLink,
		})
	}

	return repository.BatchInsert(tx, "route_infos", []string{"basin", "route_id", "start_text", "infobus_id", "pdf_link"}, values)
}
