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
		//Adds delay
		val.ArrivalTime.Time = val.ArrivalTime.Add(time.Duration(val.Delay) * time.Second)
		val.DepartureTime.Time = val.DepartureTime.Add(time.Duration(val.Delay) * time.Second)
		//Converts delay into minutes
		val.DelayMin = ConvertDelay(val.Delay)
		//Formats times
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

// rankRune assegna un peso al carattere: le lettere vengono prima dei simboli.
// Tra le lettere, ordine alfabetico; tra i simboli, ordine per valore unicode.
func rankRune(r rune) (group int, value rune) {
    if unicode.IsLetter(r) {
        return 0, unicode.ToLower(r)
    }
    return 1, r
}

// compareSuffix confronta due suffissi carattere per carattere,
// mettendo le lettere prima dei simboli (es. "/" va in fondo).
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
    // Se un suffisso è prefisso dell'altro, il più corto viene prima
    return len(ra) < len(rb)
}

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