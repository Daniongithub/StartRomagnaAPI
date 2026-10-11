package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func GetStartConfig() {
	resp, err := http.Get("https://www.startromagna.it/orari-e-percorsi-vers-2-0/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	html := string(body)

	const marker = "var SRO_CONFIG ="
	_, after, ok := strings.Cut(html, marker)
	if !ok {
		panic("marker non trovato")
	}

	// Punto di partenza: subito dopo "="
	rest := after

	//!!DA FARE MODEL CORRETTO E FATTO BENE!!

	var config map[string]any // o una struct tipizzata
	dec := json.NewDecoder(strings.NewReader(rest))
	if err := dec.Decode(&config); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", config)
}
