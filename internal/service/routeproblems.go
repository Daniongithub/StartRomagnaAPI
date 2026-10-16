package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"startromagnaapi/internal/model"
	"strings"
)

func GetRouteProblemsConfig(url string) (*model.SroConfig, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GetRouteProblems: richiesta HTTP fallita: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetRouteProblems: status code inatteso: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetRouteProblems: lettura body fallita: %w", err)
	}
	html := string(body)

	const marker = "var SRO_CONFIG ="
	_, after, ok := strings.Cut(html, marker)
	if !ok {
		return nil, fmt.Errorf("GetRouteProblems: marker %q non trovato nella pagina", marker)
	}

	// Primo livello: può essere un oggetto diretto {...} oppure una
	// stringa JSON "{...}". json.RawMessage cattura i byte così come sono,
	// poi ispezioniamo il primo carattere non-whitespace per capire quale caso è.
	var raw json.RawMessage
	dec := json.NewDecoder(strings.NewReader(after))
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("GetRouteProblems: decodifica primo livello fallita: %w", err)
	}

	jsonBytes := []byte(raw)
	trimmed := strings.TrimSpace(string(jsonBytes))

	if strings.HasPrefix(trimmed, `"`) {
		// Caso stringa: è JSON-escaped, va "srotolato" un livello in più.
		var inner string
		if err := json.Unmarshal(jsonBytes, &inner); err != nil {
			return nil, fmt.Errorf("GetRouteProblems: unescape stringa fallito: %w", err)
		}
		jsonBytes = []byte(inner)
	}

	var config model.SroConfig
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, fmt.Errorf("GetRouteProblems: unmarshal config fallito: %w", err)
	}

	return &config, nil
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
