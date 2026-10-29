package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"startromagnaapi/internal/auth"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"strings"
)

func ScrapeRouteInfo(url string) (*model.RouteInfoRaw, error) {
	resp, err := auth.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ScrapeRouteInfo: Unexpected HTTP status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: Failed to read the response body: %w", err)
	}
	html := string(body)

	const marker = "var SRO_CONFIG ="
	_, after, ok := strings.Cut(html, marker)
	if !ok {
		return nil, fmt.Errorf("ScrapeRouteInfo: Marker %q was not found in the page", marker)
	}

	// The outer value can be an object {...} or an escaped JSON string.
	// json.RawMessage preserves the original bytes so the first
	// non-whitespace character identifies which representation was used.
	var raw json.RawMessage
	dec := json.NewDecoder(strings.NewReader(after))
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: Failed to decode the outer JSON value: %w", err)
	}

	jsonBytes := []byte(raw)
	trimmed := strings.TrimSpace(string(jsonBytes))

	if strings.HasPrefix(trimmed, `"`) {
		// Decode one extra layer when the value is an escaped JSON string.
		var inner string
		if err := json.Unmarshal(jsonBytes, &inner); err != nil {
			return nil, fmt.Errorf("ScrapeRouteInfo: Failed to unescape the JSON string: %w", err)
		}
		jsonBytes = []byte(inner)
	}

	var config model.RouteInfoRaw
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: Failed to decode the route configuration: %w", err)
	}

	return &config, nil
}

func ProcessRouteInfo(infos model.RouteInfoRaw) []model.RouteInfo {
	knownRoutes := static.GetRoutes()
	routesMap := make(map[[2]string]bool)
	for _, val := range knownRoutes {
		routesMap[[2]string{val.Basin, val.Route_id}] = true
	}
	var processed []model.RouteInfo
	for area, slice := range infos.Areas {
		for _, val := range slice {
			var basin string
			if area == "ravenna" {
				basin = "RA"
			}
			if area == "cesenatico" {
				basin = "FC"
			}
			if area == "cesena" {
				basin = "FC"
			}
			if area == "forli" {
				basin = "FC"
			}
			if area == "extra" {
				basin = "FC"
			}
			if area == "rimini" {
				basin = "RN"
			}
			if val.Route == "1-2CO" {
				val.Route = "1-2C"
			}
			if val.Route == "F126p" {
				val.Route = "126p"
			}
			if val.Route == "F131p" {
				val.Route = "131p"
			}
			if val.Route == "F132p" {
				val.Route = "132p"
			}
			if val.Route == "Metromare" {
				val.Route = "9311"
			}
			_, ok := routesMap[[2]string{basin, val.Route}]
			if ok {
				processed = append(processed, model.RouteInfo{
					Basin:     basin,
					RouteId:   val.Route,
					StartText: val.Title,
					InfobusId: val.InfoBusId,
					PDFLink:   val.Pdf,
				})
			}
		}
	}

	return processed
}
