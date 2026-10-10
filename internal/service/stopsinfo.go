package service

import (
	"regexp"
	"strings"

	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/utils"
)

type lineCategory int

const (
	categoryUrban lineCategory = iota
	categorySuburban
	categoryExurban
	categoryShuttle
	categorySchool
)

var (
	// 1xx / 2xx, with optional letter suffix (e.g. "132p")
	exurbanTokenRegex = regexp.MustCompile(`^[12]\d{2}[A-Za-z]*$`)
	// 9x with optional letter suffix (e.g. "94", "94A")
	suburbanTokenRegex = regexp.MustCompile(`^(9\d)[A-Za-z]?$`)
	// School lines start with "S" (e.g. "S1 Forlì")
	schoolTokenRegex = regexp.MustCompile(`^[Ss]`)
)

var (
	ravennaShuttleTokens = map[string]struct{}{
		"65": {}, "66": {}, "67": {}, "parkibus": {},
	}
	forliUrbanNinetyLines = map[string]struct{}{
		"91": {}, "93": {},
	}
)

// firstToken returns the first whitespace-separated word of a line name.
func firstToken(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func classifyLine(basin, lineName string) lineCategory {
	token := firstToken(lineName)

	switch strings.ToUpper(basin) {
	case "RA":
		if _, ok := ravennaShuttleTokens[strings.ToLower(token)]; ok {
			return categoryShuttle
		}
		if exurbanTokenRegex.MatchString(token) {
			return categoryExurban
		}
		return categoryUrban

	case "FC":
		if schoolTokenRegex.MatchString(token) {
			return categorySchool
		}
		if exurbanTokenRegex.MatchString(token) {
			return categoryExurban
		}
		if m := suburbanTokenRegex.FindStringSubmatch(token); m != nil {
			if _, isUrban := forliUrbanNinetyLines[m[1]]; !isUrban {
				return categorySuburban
			}
		}
		return categoryUrban

	default: // "RN" and any other basin: no distinction
		return categoryUrban
	}
}

func ProcessStopsInfo(stopCode, basin string) model.StopInfo {
	rows := static.GetLinesForStop(stopCode, basin)

	seen := make(map[string]struct{}, len(rows))
	buckets := map[lineCategory][]model.Line{
		categoryUrban:    make([]model.Line, 0, len(rows)),
		categoryExurban:  make([]model.Line, 0, len(rows)),
		categorySuburban: nil,
		categoryShuttle:  nil,
		categorySchool:   nil,
	}

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

		// Deduplication
		key := line.RouteId + "|" + line.Line
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		// Separation
		category := classifyLine(basin, line.Line)
		buckets[category] = append(buckets[category], line)
	}

	for _, lines := range buckets {
		utils.SortLines(lines)
	}

	return model.StopInfo{
		Urban:    buckets[categoryUrban],
		Suburban: buckets[categorySuburban],
		Exurban:  buckets[categoryExurban],
		Shuttle:  buckets[categoryShuttle],
		School:   buckets[categorySchool],
	}
}
