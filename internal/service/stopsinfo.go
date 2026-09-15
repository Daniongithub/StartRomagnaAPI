package service

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/utils"
)

func ProcessStopsInfo(stopCode, basin string) model.StopInfo {
    var stopInfo model.StopInfo

    rows := static.GetLinesForStop(stopCode, basin)

    seen := make(map[string]struct{})
    lines := make([]model.Line, 0, len(rows))

    for _, row := range rows {
        line := model.Line{
            RouteId:      row.RouteID,
            OfficialLine: row.OfficialLine,
        }
        if row.DispLine != nil {
            line.Line = *row.DispLine
        } else {
            line.Line = row.OfficialLine
        }

        //Deduplication
        key := line.RouteId + "|" + line.Line
        if _, exists := seen[key]; exists {
            continue
        }
        seen[key] = struct{}{}

        lines = append(lines, line)
    }

	//Sort by numeric part (letter only stuff goes at the bottom)
	utils.SortLines(lines)
    stopInfo.Lines = lines
    return stopInfo
}