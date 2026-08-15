package main

import (
	"StartRomagnaAPI/config"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.LoadConf()

	mux := http.NewServeMux()

	//Listen on port and start API
	fmt.Println("Server started on port " + config.PORT)
	log.Print(http.ListenAndServe(config.PORT, mux))
}
