package repository

import (
	"database/sql"
	"fmt"
	"startromagnaapi/config"
	"startromagnaapi/internal/model"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	DB_STATIC *sqlx.DB
	DB_RT     *sqlx.DB
	DB_MEZZI  *sqlx.DB
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

	config.IS_PRIMARY, err = isPrimary(DB_RT)
	if err != nil {
		fmt.Println("InitRT errore determinazione DB principale:", err)
	}
}

func InitMezzi() {
	DB_MEZZI, err = newDB(config.DB_HOST, config.DB_PORT, config.DB_USERNAME, config.DB_PASSWORD, "ertpl_mezzi")
	if err != nil {
		fmt.Println("InitMezzi errore di connessione al database del GTFS dinamico:", err)
	}

	config.IS_PRIMARY, err = isPrimary(DB_MEZZI)
	if err != nil {
		fmt.Println("InitMezzi errore determinazione DB principale:", err)
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

func BatchInsertTX(db *sqlx.Tx, table string, columns []string, new [][]any) error {
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

func GetDBStatus(db *sql.DB) (model.DBStatus, error) {
	var readOnly string
	if err := db.QueryRow("SHOW GLOBAL VARIABLES LIKE 'read_only'").Scan(new(string), &readOnly); err != nil {
		return model.DBStatus{}, err
	}

	var gtidPos string
	if err := db.QueryRow("SELECT @@GLOBAL.gtid_current_pos").Scan(&gtidPos); err != nil {
		return model.DBStatus{}, err
	}

	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return model.DBStatus{}, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	hasSlaveRow := rows.Next()

	status := model.DBStatus{
		ReadOnly:     readOnly == "ON",
		GtidPosition: gtidPos,
	}

	if !hasSlaveRow {
		status.Role = "master"
		return status, nil
	}

	// scan dinamico perché SHOW SLAVE STATUS ha decine di colonne
	// e cambia leggermente tra versioni MariaDB
	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]any, len(cols))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	if err := rows.Scan(scanArgs...); err != nil {
		return model.DBStatus{}, err
	}
	col := func(name string) string {
		for i, c := range cols {
			if c == name {
				return string(values[i])
			}
		}
		return ""
	}

	ioRunning := col("Slave_IO_Running") == "Yes"
	sqlRunning := col("Slave_SQL_Running") == "Yes"

	var lag *int
	if v := col("Seconds_Behind_Master"); v != "" {
		n, _ := strconv.Atoi(v)
		lag = &n
	}

	rep := &model.ReplicationStatus{
		MasterHost: col("Master_Host"),
		IORunning:  ioRunning,
		SQLRunning: sqlRunning,
		LagSeconds: lag,
		CaughtUp:   lag != nil && *lag == 0,
	}
	status.Replication = rep

	switch {
	case !ioRunning || !sqlRunning:
		status.Role = "broken"
	case lag == nil || *lag > 0:
		status.Role = "resyncing"
	default:
		status.Role = "replica"
	}

	return status, nil
}
