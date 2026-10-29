package service

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/realtime"
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
				p := splitLine(it.Line)
				if _, exists := cityGroups[p.city]; !exists {
					cityOrder = append(cityOrder, p.city)
				}
				cityGroups[p.city] = append(cityGroups[p.city], it)
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
					pi := splitLine(g[i].Line)
					pj := splitLine(g[j].Line)
					if pi.num != pj.num {
						return pi.num < pj.num
					}
					return pi.variant < pj.variant
				})
				result = append(result, g...)
			}
		} else {
			sort.SliceStable(basinItems, func(i, j int) bool {
				pi := splitLine(basinItems[i].Line)
				pj := splitLine(basinItems[j].Line)

				if pi.num != pj.num {
					return pi.num < pj.num
				}
				return pi.variant < pj.variant
			})

			result = append(result, basinItems...)
		}
	}

	return result
}
