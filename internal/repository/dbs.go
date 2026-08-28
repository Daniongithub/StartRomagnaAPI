package repository

import (
	"StartRomagnaAPI/config"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	DB_STATIC *sqlx.DB
	DB_RT     *sqlx.DB
	err       error
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

func isPrimary(db *sqlx.DB) (bool, error) {
	var varName, value string
	row := db.QueryRow("SHOW VARIABLES LIKE 'read_only'")
	if err := row.Scan(&varName, &value); err != nil {
		return false, fmt.Errorf("lettura read_only: %w", err)
	}
	return value == "OFF" || value == "0", nil

}

func InitStatic() {
	DB_STATIC, err = newDB(config.DB_HOST, config.DB_PORT, config.DB_USERNAME, config.DB_PASSWORD, "start_gtfs_static")
	if err != nil {
		fmt.Println("InitStatic errore di connessione al database del GTFS statico:", err)
	}

	config.IS_PRIMARY, err = isPrimary(DB_STATIC)
	if err != nil {
		fmt.Println("InitStatic errore determinazione DB principale:", err)
	}
}

func InitRT() {
	DB_RT, err = newDB(config.DB_HOST, config.DB_PORT, config.DB_USERNAME, config.DB_PASSWORD, "start_gtfs_rt")
	if err != nil {
		fmt.Println("InitRT errore di connessione al database del GTFS dinamico:", err)
	}

	config.IS_PRIMARY, err = isPrimary(DB_STATIC)
	if err != nil {
		fmt.Println("InitRT errore determinazione DB principale:", err)
	}
}

const insertBatchSize = 5000

func BatchInsert(db *sqlx.DB, table string, columns []string, new [][]any) error {
	if len(new) == 0 {
		return nil
	}

	for start := 0; start < len(new); start += insertBatchSize {
		end := min(start+insertBatchSize, len(new))

		q := squirrel.Insert(table).
			Columns(columns...)

		for _, row := range new[start:end] {
			q = q.Values(row...)
		}

		query, args, err := q.ToSql()
		if err != nil {
			return fmt.Errorf("build insert query: %w", err)
		}

		if _, err := db.Exec(query, args...); err != nil {
			return fmt.Errorf("execute insert into %s: %w", table, err)
		}
	}

	return nil
}
