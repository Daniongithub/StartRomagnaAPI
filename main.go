package main

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/handler"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.LoadConf()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthcheckHandler)

	//Listen on port and start API
	fmt.Println("Server started on port " + config.PORT)
	log.Print(http.ListenAndServe(config.PORT, mux))
}
