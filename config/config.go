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

	ALLOWED_ORIGINS []string
)

func LoadConf() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Errore: nessun file .env trovato. Il programma non può continuare.", err)
	}

	START_GTFS_ROOT = os.Getenv("START_GTFS_ROOT")
	START_GTFS_RT_ROOT = os.Getenv("START_GTFS_RT_ROOT")
	WEB_AUTH_USER = os.Getenv("WEB_AUTH_USER")
	WEB_AUTH_PASSWORD = os.Getenv("WEB_AUTH_PASSWORD")

	DB_HOST = os.Getenv("DB_HOST")
	DB_PORT, _ = strconv.Atoi(os.Getenv("DB_PORT"))
	DB_USERNAME = os.Getenv("DB_USERNAME")
	DB_PASSWORD = os.Getenv("DB_PASSWORD")

	PORT = os.Getenv("PORT")

	ALLOWED_ORIGINS = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
}
