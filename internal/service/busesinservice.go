package service

import (
	"errors"
	"sort"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
	"startromagnaapi/internal/utils"
)

var NoBusesInService = errors.New("No buses in service")

func ProcessBusesInService() ([]model.BusInService, error) {
	buses := realtime.GetBuses()

	if len(buses) == 0 {
		return nil, NoBusesInService
	}

	enrich := newBusEnricher()
	for i := range buses {
		enrich.enrich(&buses[i])
	}
	buses = groupAndSort(buses)

	return buses, nil
}

// SORT AND GROUP BY LINE

func groupAndSort(items []model.BusInService) []model.BusInService {
	// Group by basin in order of appearance.
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

		if basin == "FC" {
			// Group lines by city within the basin.
			cityOrder := []string{}
			cityGroups := make(map[string][]model.BusInService)

			for _, it := range basinItems {
				p := utils.SplitLine(it.Line)
				if _, exists := cityGroups[p.City]; !exists {
					cityOrder = append(cityOrder, p.City)
				}
				cityGroups[p.City] = append(cityGroups[p.City], it)
			}

			// Place lines without a city suffix first, then sort cities alphabetically.
			sort.SliceStable(cityOrder, func(i, j int) bool {
				if cityOrder[i] != "" {
					return true
				}
				if cityOrder[j] != "" {
					return false
				}
				return cityOrder[i] < cityOrder[j]
			})

			// Sort by line number, then by variant (1, 1B, 2, 3, 4B...).
			for _, city := range cityOrder {
				g := cityGroups[city]
				sort.SliceStable(g, func(i, j int) bool {
					pi := utils.SplitLine(g[i].Line)
					pj := utils.SplitLine(g[j].Line)
					if pi.Num != pj.Num {
						return pi.Num < pj.Num
					}
					return pi.Variant < pj.Variant
				})
				result = append(result, g...)
			}
		} else {
			sort.SliceStable(basinItems, func(i, j int) bool {
				pi := utils.SplitLine(basinItems[i].Line)
				pj := utils.SplitLine(basinItems[j].Line)

				if pi.Num != pj.Num {
					return pi.Num < pj.Num
				}
				return pi.Variant < pj.Variant
			})

			result = append(result, basinItems...)
		}
	}

	return result
}
