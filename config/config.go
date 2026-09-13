package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	START_GTFS_ROOT    string
	START_GTFS_RT_ROOT string
	WEB_AUTH_USER      string
	WEB_AUTH_PASSWORD  string

	DB_HOST     string
	DB_PORT     int
	DB_USERNAME string
	DB_PASSWORD string

	PORT string

	IS_PRIMARY      bool
	ALLOWED_ORIGINS []string

	ARRIVALS_LOAD_INTERVAL int
	ARRIVALS_DELAY_BUFFER  int
)

func LoadConf() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Errore: nessun file .env trovato. Il programma non può continuare.", err)
	}

	getRequired := func(key string) string {
		value, exists := os.LookupEnv(key)
		if !exists || strings.TrimSpace(value) == "" {
			log.Fatalf("Parametro obbligatorio mancante nel file .env: %s", key)
		}
		return value
	}

	getRequiredInt := func(key string) int {
		value := getRequired(key)

		result, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("Parametro %s non è un numero valido: %q", key, value)
		}

		return result
	}

	START_GTFS_ROOT = getRequired("START_GTFS_ROOT")
	START_GTFS_RT_ROOT = getRequired("START_GTFS_RT_ROOT")
	WEB_AUTH_USER = getRequired("WEB_AUTH_USER")
	WEB_AUTH_PASSWORD = getRequired("WEB_AUTH_PASSWORD")

	DB_HOST = getRequired("DB_HOST")
	DB_PORT = getRequiredInt("DB_PORT")
	DB_USERNAME = getRequired("DB_USERNAME")
	DB_PASSWORD = getRequired("DB_PASSWORD")

	PORT = getRequired("PORT")

	ALLOWED_ORIGINS = strings.Split(getRequired("ALLOWED_ORIGINS"), ",")

	ARRIVALS_LOAD_INTERVAL = getRequiredInt("ARRIVALS_LOAD_INTERVAL")
	ARRIVALS_DELAY_BUFFER = getRequiredInt("ARRIVALS_DELAY_BUFFER")
}
