package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
)

func ProcessCorsesopp(basin string) model.CorseSoppStop {
	firstStop := realtime.GetFirstStop("8003_A-123")

	return firstStop
}
