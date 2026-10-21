package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
)

var (
	START_GTFS_ROOT    string
	START_GTFS_RT_ROOT string
	WEB_AUTH_USER      string
	WEB_AUTH_PASSWORD  string

	OIDC_ISSUER_URL     string
	OIDC_CLIENT_ID      string
	OIDC_ALLOWED_GROUPS []string
	ENABLE_AUTH         bool

	DB_HOST          string
	DB_PORT          int
	DB_USERNAME      string
	DB_PASSWORD      string
	DB_ORCH_USERNAME string
	DB_ORCH_PASSWORD string
	REPLICA_USERNAME string
	REPLICA_PASSWORD string
	ARBITER_SECRET   string

	PORT string

	IS_PRIMARY      atomic.Bool
	ALLOWED_ORIGINS []string

	ARRIVALS_LOAD_INTERVAL int
	ARRIVALS_DELAY_BUFFER  int
	SCHEDULER              gocron.Scheduler
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

	getRequiredBool := func(key string) bool {
		value, exists := os.LookupEnv(key)
		if !exists || strings.TrimSpace(value) == "" {
			log.Fatalf("Parametro obbligatorio mancante nel file .env: %s", key)
		} else if value == "true" {
			return true
		}
		return false
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

	//OAuth config
	OIDC_ISSUER_URL = getRequired("OIDC_ISSUER_URL")
	OIDC_CLIENT_ID = getRequired("OIDC_CLIENT_ID")
	OIDC_ALLOWED_GROUPS = strings.Split(getRequired("OIDC_ALLOWED_GROUPS"), ",")
	ENABLE_AUTH = getRequiredBool("ENABLE_AUTH")

	DB_HOST = getRequired("DB_HOST")
	DB_PORT = getRequiredInt("DB_PORT")
	DB_USERNAME = getRequired("DB_USERNAME")
	DB_PASSWORD = getRequired("DB_PASSWORD")
	DB_ORCH_USERNAME = getRequired("DB_ORCH_USERNAME")
	DB_ORCH_PASSWORD = getRequired("DB_ORCH_PASSWORD")
	ARBITER_SECRET = getRequired("ARBITER_SECRET")

	PORT = getRequired("PORT")

	ALLOWED_ORIGINS = strings.Split(getRequired("ALLOWED_ORIGINS"), ",")

	ARRIVALS_LOAD_INTERVAL = getRequiredInt("ARRIVALS_LOAD_INTERVAL")
	ARRIVALS_DELAY_BUFFER = getRequiredInt("ARRIVALS_DELAY_BUFFER")
}
