package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/utils"
)

func ProcessNextstops(tripId string) model.NextStops {
	nextstops := realtime.GetNextStops(tripId)

	utils.FixStopWDel(nextstops.Stops)

	return nextstops
}
