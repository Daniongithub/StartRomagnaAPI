package utils

import (
	"regexp"
	"sort"
	"startromagnaapi/internal/model"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func ConvertDelay(delay int) int {
	mins := delay / 60
	rest := delay % 60
	if rest >= 30 {
		mins++
	} else if rest <= -30 {
		mins--
	}
	return mins
}

func FixStopWDel(stops []model.StopWDel) {
	for idx := range stops {
		val := &stops[idx]
		// Apply the delay.
		val.ArrivalTime.Time = val.ArrivalTime.Add(time.Duration(val.Delay) * time.Second)
		val.DepartureTime.Time = val.DepartureTime.Add(time.Duration(val.Delay) * time.Second)
		// Convert the delay to minutes.
		val.DelayMin = ConvertDelay(val.Delay)
		// Format the arrival and departure times.
		val.ArrivalTimeStr = val.ArrivalTime.Format("15:04:05")
		val.DepartureTimeStr = val.DepartureTime.Format("15:04:05")
	}
}

var lineNumberRegex = regexp.MustCompile(`^(\d+)(.*)$`)

func parseLine(line string) (number int, suffix string, hasNumber bool) {
	matches := lineNumberRegex.FindStringSubmatch(strings.TrimSpace(line))
	if matches == nil {
		return 0, line, false
	}
	n, _ := strconv.Atoi(matches[1])
	return n, matches[2], true
}

func rankRune(r rune) (group int, value rune) {
	if unicode.IsLetter(r) {
		return 0, unicode.ToLower(r)
	}
	return 1, r
}

func compareSuffix(a, b string) bool {
	ra, rb := []rune(a), []rune(b)
	for i := 0; i < len(ra) && i < len(rb); i++ {
		gA, vA := rankRune(ra[i])
		gB, vB := rankRune(rb[i])
		if gA != gB {
			return gA < gB
		}
		if vA != vB {
			return vA < vB
		}
	}
	return len(ra) < len(rb)
}

func lineValue(l model.RoutesWOffLine) string {
	if l.DispLine != nil && *l.DispLine != "" {
		return *l.DispLine
	}
	return l.OfficialLine
}

// Sort lines by their numeric component.
func SortLines(lines []model.Line) {
	sort.SliceStable(lines, func(i, j int) bool {
		numI, sufI, hasNumI := parseLine(lines[i].Line)
		numJ, sufJ, hasNumJ := parseLine(lines[j].Line)

		if hasNumI != hasNumJ {
			return hasNumI
		}

		if !hasNumI {
			return compareSuffix(sufI, sufJ)
		}

		if numI != numJ {
			return numI < numJ
		}
		return compareSuffix(sufI, sufJ)
	})
}

var reLine = regexp.MustCompile(`^(\d+)([A-Za-z]*)\s*(.*)$`)

// Sort routes by city for FC, then by line number and variant.
func SortRoutes(lines []model.RoutesWOffLine) []model.RoutesWOffLine {
	if len(lines) < 2 {
		return nil
	}
	var result []model.RoutesWOffLine
	if lines[0].Basin == "FC" {
		// Group lines by city within the basin.
		cityOrder := []string{}
		cityGroups := make(map[string][]model.RoutesWOffLine)

		for _, it := range lines {
			p := SplitLine(lineValue(it))
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
				pi := SplitLine(lineValue(g[i]))
				pj := SplitLine(lineValue(g[j]))
				if pi.Num != pj.Num {
					return pi.Num < pj.Num
				}
				return pi.Variant < pj.Variant
			})
			result = append(result, g...)
		}
	} else {
		sort.SliceStable(lines, func(i, j int) bool {
			pi := SplitLine(lineValue(lines[i]))
			pj := SplitLine(lineValue(lines[j]))

			if pi.Num != pj.Num {
				return pi.Num < pj.Num
			}
			return pi.Variant < pj.Variant
		})

		result = append(result, lines...)
	}

	return result
}

type lineParts struct {
	Num     int
	Variant string
	City    string
}

func SplitLine(line string) lineParts {
	matches := reLine.FindStringSubmatch(strings.TrimSpace(line))
	if matches == nil {
		return lineParts{Num: 1<<31 - 1, Variant: line, City: ""}
	}
	n, _ := strconv.Atoi(matches[1])
	return lineParts{
		Num:     n,
		Variant: matches[2],
		City:    strings.TrimSpace(matches[3]),
	}
}

func FindRouteFromID(arr []model.RoutesResult, routeId string) *model.RoutesResult {
	for _, val := range arr {
		if val.Route_id == routeId {
			return &val
		}
	}
	return nil
}
