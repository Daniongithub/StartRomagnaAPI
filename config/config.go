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

	OIDC_ISSUER_URL     string
	OIDC_CLIENT_ID      string
	OIDC_ALLOWED_GROUPS []string
	ENABLE_AUTH         bool

	DB_HOST     string
	DB_PORT     int
	DB_USERNAME string
	DB_PASSWORD string

	PORT string

	IS_PRIMARY bool

	ARRIVALS_LOAD_INTERVAL int
	ARRIVALS_DELAY_BUFFER  int
)

const (
	FEED_URL       = "https://www.startromagna.it/infobus/feed/"
	ROUTE_INFO_URL = "https://www.startromagna.it/orari-e-percorsi-vers-2-0/"
)

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Missing .env file. The application cannot continue.", err)
	}

	getRequired := func(key string) string {
		value, exists := os.LookupEnv(key)
		if !exists || strings.TrimSpace(value) == "" {
			log.Fatalf("Missing required environment variable: %s", key)
		}
		return value
	}

	getRequiredBool := func(key string) bool {
		value, err := strconv.ParseBool(strings.TrimSpace(getRequired(key)))
		if err != nil {
			log.Fatalf("Invalid boolean value for %s", key)
		}
		return value
	}

	getRequiredInt := func(key string) int {
		value := getRequired(key)

		result, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("Invalid integer value for %s: %q", key, value)
		}

		return result
	}

	START_GTFS_ROOT = getRequired("START_GTFS_ROOT")
	WEB_AUTH_USER = getRequired("WEB_AUTH_USER")
	WEB_AUTH_PASSWORD = getRequired("WEB_AUTH_PASSWORD")

	DB_HOST = getRequired("DB_HOST")
	DB_PORT = getRequiredInt("DB_PORT")
	DB_USERNAME = getRequired("DB_USERNAME")
	DB_PASSWORD = getRequired("DB_PASSWORD")
	if DB_PORT < 1 || DB_PORT > 65535 {
		log.Fatal("DB_PORT must be between 1 and 65535")
	}
	IS_PRIMARY = getRequiredBool("IS_PRIMARY")

	START_GTFS_RT_ROOT = getRequired("START_GTFS_RT_ROOT")

	// OAuth configuration.
	ENABLE_AUTH = getRequiredBool("ENABLE_AUTH")
	if ENABLE_AUTH {
		OIDC_ISSUER_URL = getRequired("OIDC_ISSUER_URL")
		OIDC_CLIENT_ID = getRequired("OIDC_CLIENT_ID")
		OIDC_ALLOWED_GROUPS = strings.Split(getRequired("OIDC_ALLOWED_GROUPS"), ",")
	}

	PORT = getRequired("PORT")

	ARRIVALS_LOAD_INTERVAL = getRequiredInt("ARRIVALS_LOAD_INTERVAL")
	ARRIVALS_DELAY_BUFFER = getRequiredInt("ARRIVALS_DELAY_BUFFER")
	if ARRIVALS_LOAD_INTERVAL <= 0 || ARRIVALS_DELAY_BUFFER < 0 {
		log.Fatal("Invalid arrival intervals")
	}
}
