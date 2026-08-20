package repository

import (
	"StartRomagnaAPI/config"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	DB_CONTENT *sqlx.DB
	err        error
)

func newDB(host string, port int, user, password, dbname string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, dbname)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func InitContent() {
	DB_CONTENT, err = newDB(config.DB_HOST, config.DB_PORT, config.DB_USERNAME, config.DB_PASSWORD, "start_gtfs_content")
	if err != nil {
		fmt.Println("InitContent errore di connessione al database contenuti:", err)
	}
}
