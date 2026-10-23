package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository/static"
	"strings"
)

func ScrapeRouteInfo(url string) (*model.RouteInfoRaw, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: richiesta HTTP fallita: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ScrapeRouteInfo: status code inatteso: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: lettura body fallita: %w", err)
	}
	html := string(body)

	const marker = "var SRO_CONFIG ="
	_, after, ok := strings.Cut(html, marker)
	if !ok {
		return nil, fmt.Errorf("ScrapeRouteInfo: marker %q non trovato nella pagina", marker)
	}

	// Primo livello: può essere un oggetto diretto {...} oppure una
	// stringa JSON "{...}". json.RawMessage cattura i byte così come sono,
	// poi ispezioniamo il primo carattere non-whitespace per capire quale caso è.
	var raw json.RawMessage
	dec := json.NewDecoder(strings.NewReader(after))
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: decodifica primo livello fallita: %w", err)
	}

	jsonBytes := []byte(raw)
	trimmed := strings.TrimSpace(string(jsonBytes))

	if strings.HasPrefix(trimmed, `"`) {
		// Caso stringa: è JSON-escaped, va "srotolato" un livello in più.
		var inner string
		if err := json.Unmarshal(jsonBytes, &inner); err != nil {
			return nil, fmt.Errorf("ScrapeRouteInfo: unescape stringa fallito: %w", err)
		}
		jsonBytes = []byte(inner)
	}

	var config model.RouteInfoRaw
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, fmt.Errorf("ScrapeRouteInfo: unmarshal config fallito: %w", err)
	}

	return &config, nil
}

func ProcessRouteInfo(infos model.RouteInfoRaw) []model.RouteInfo {
	knownRoutes := static.GetRoutes()
	routesMap := make(map[string]bool)
	for _, val := range knownRoutes {
		routesMap[val.Route_id] = true
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
			_, ok := routesMap[val.Route]
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

/*

PROSSIMI STEP: INSERIRE DATI NEL DB OGNI MEZZANOTTE
GESTIRE CONVERSIONE LINEA CHE MOSTRIAMO CON LINEA DEL JSON ("1 Cesena" -> CE01, oppure tenere nei JSON anche official_line), FARE JOIN
LINEA 4 RAVENNA --> id 612 secondo il JSON

Esempio:

curl -X POST 'https://www.startromagna.it/wp-admin/admin-ajax.php' -H 'Content-Type: application/x-www-form-urlencoded;charset=UTF-8' --data-urlencode 'action=get_infobus_linea' --data-urlencode 'idPost=612'
NON RICHIEDE COOKIE O AUTENTICAZIONE!

SCRAPARE HTML PER FARE JSON (ci pensa il Dani)
*/
