package service

import (
	"regexp"
	"sort"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/mezzi"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/repository/static"
	"strconv"
	"strings"
)

func ProcessBusesInService() []model.BusInService {
	buses := realtime.GetBuses()

	for idx := range buses {
		val := &buses[idx]
		val.NextStop = realtime.GetFirstStop(val.TripId)
		val.OfficialLine = static.GetRouteNamefromID(val.Basin, val.RouteId)
		headsign := static.GetHeadsignsByID(val.ShapeId)
		if headsign != nil {
			if headsign.DispLine != nil {
				val.Line = *headsign.DispLine
			} else {
				val.Line = val.OfficialLine
			}
			if headsign.DispDest != nil {
				val.Destination = *headsign.DispDest
			} else {
				val.Destination = strings.ToUpper(static.GetTerminusName(val.TripId))
			}
		}
		//Model and stuff
		if mezzi.GetVehicleInServiceByID(val.Vehicle) != nil {
			val.VehicleInfo = *mezzi.GetVehicleInServiceByID(val.Vehicle)
		} else {
			val.VehicleInfo = model.VehicleInService{
				Number: val.Vehicle,
			}
		}
		//METE model and stuff
		if mezzi.GetMeteInServiceByID(val.Vehicle) != nil {
			val.VehicleInfo = *mezzi.GetMeteInServiceByID(val.Vehicle)
		} else {
			val.VehicleInfo = model.VehicleInService{
				Number: val.Vehicle,
			}
		}
	}
	buses = groupAndSort(buses)

	return buses
}

// SORT AND GROUP BY LINE

var reLine = regexp.MustCompile(`^(\d+)([A-Za-z]*)\s*(.*)$`)

type lineParts struct {
	num     int
	variant string
	city    string
}

func splitLine(line string) lineParts {
	matches := reLine.FindStringSubmatch(strings.TrimSpace(line))
	if matches == nil {
		return lineParts{num: 1<<31 - 1, variant: "", city: line}
	}
	n, _ := strconv.Atoi(matches[1])
	return lineParts{
		num:     n,
		variant: matches[2],
		city:    strings.TrimSpace(matches[3]),
	}
}

func groupAndSort(items []model.BusInService) []model.BusInService {
	// --- Livello 1: raggruppa per Basin (ordine di apparizione) ---
	basinOrder := []string{}
	basinGroups := make(map[string][]model.BusInService)

	for _, it := range items {
		if _, exists := basinGroups[it.Basin]; !exists {
			basinOrder = append(basinOrder, it.Basin)
		}
		basinGroups[it.Basin] = append(basinGroups[it.Basin], it)
	}
	
	result := make([]model.BusInService, 0, len(items))

	for _, basin := range basinOrder {
		basinItems := basinGroups[basin]

		// --- Livello 2: raggruppa per "città" dentro il basin ---
		cityOrder := []string{}
		cityGroups := make(map[string][]model.BusInService)

		for _, it := range basinItems {
			p := splitLine(it.Line)
			if _, exists := cityGroups[p.city]; !exists {
				cityOrder = append(cityOrder, p.city)
			}
			cityGroups[p.city] = append(cityGroups[p.city], it)
		}

		// Città senza suffisso ("") prima, poi le altre in ordine alfabetico
		sort.SliceStable(cityOrder, func(i, j int) bool {
			if cityOrder[i] != "" {
				return true
			}
			if cityOrder[j] != "" {
				return false
			}
			return cityOrder[i] < cityOrder[j]
		})

		// --- Livello 3: ordina per numero, poi per variante (1, 1B, 2, 3, 4B...) ---
		for _, city := range cityOrder {
			g := cityGroups[city]
			sort.SliceStable(g, func(i, j int) bool {
				pi := splitLine(g[i].Line)
				pj := splitLine(g[j].Line)
				if pi.num != pj.num {
					return pi.num < pj.num
				}
				return pi.variant < pj.variant
			})
			result = append(result, g...)
		}
	}

	return result
}
