package handler

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/auth"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	message := "API is healthy and running on port " + config.PORT

	w.Write([]byte(message))
}

func GTFSHandler(w http.ResponseWriter, r *http.Request) {
	url := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Ravenna.zip"

	req, err := auth.BasicAuth("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	reader := bytes.NewReader(body)

	zipReader, err := zip.NewReader(reader, int64(len(body)))
	if err != nil {
		return
	}

	for _, file := range zipReader.File {
		fmt.Println(file.Name)
	}
}
