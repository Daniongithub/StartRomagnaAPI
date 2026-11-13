package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"startromagnaapi/config"
	"startromagnaapi/internal/model"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	DB_STATIC *sqlx.DB
	DB_RT     *sqlx.DB
	DB_MEZZI  *sqlx.DB
)

func newDB(host string, port int, user, password, dbname string) (*sqlx.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User, cfg.Passwd = user, password
	cfg.Net, cfg.Addr, cfg.DBName = "tcp", net.JoinHostPort(host, strconv.Itoa(port)), dbname
	cfg.ParseTime = true
	cfg.MultiStatements = true
	cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 10*time.Second, 30*time.Second, 30*time.Second
	db, err := sqlx.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := pingDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

// Ping the server before allowing startup to continue. Do not retry on failure.
func pingDB(db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return err
	}
	return nil
}

func initDB(name, user, password string) *sqlx.DB {
	db, err := newDB(config.DB_HOST, config.DB_PORT, user, password, name)
	if err != nil {
		log.Fatalf("Failed to connect to database %s: %v", name, err)
	}
	return db
}

func InitStatic() { DB_STATIC = initDB("start_gtfs_static", config.DB_USERNAME, config.DB_PASSWORD) }
func InitRT()     { DB_RT = initDB("start_gtfs_rt", config.DB_USERNAME, config.DB_PASSWORD) }
func InitMezzi()  { DB_MEZZI = initDB("ertpl_mezzi", config.DB_USERNAME, config.DB_PASSWORD) }

const insertBatchSize = 5000

// Execer is implemented by both sqlx.DB and sqlx.Tx.
type Execer interface {
	Exec(string, ...any) (sql.Result, error)
}

func BatchInsert(db Execer, table string, columns []string, rows [][]any) error {
	if len(columns) == 0 {
		return fmt.Errorf("Insert %s: no columns", table)
	}
	// MySQL limits prepared statements to 65535 placeholders.
	batchSize := min(insertBatchSize, 65535/len(columns))
	if batchSize == 0 {
		return fmt.Errorf("Insert %s: too many columns", table)
	}
	for start := 0; start < len(rows); start += batchSize {
		end := min(start+batchSize, len(rows))
		q := squirrel.Insert(table).Columns(columns...)
		for _, row := range rows[start:end] {
			if len(row) != len(columns) {
				return fmt.Errorf("Insert %s: expected %d values, got %d", table, len(columns), len(row))
			}
			q = q.Values(row...)
		}
		query, args, err := q.ToSql()
		if err != nil {
			return fmt.Errorf("Build insert %s: %w", table, err)
		}
		if _, err := db.Exec(query, args...); err != nil {
			return fmt.Errorf("Insert %s: %w", table, err)
		}
	}
	return nil
}

func GetDBStatus(db *sql.DB) (model.DBStatus, error) {
	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return model.DBStatus{}, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	hasSlaveRow := rows.Next()

	status := model.DBStatus{}

	if !hasSlaveRow {
		status.Role = "master"
		return status, nil
	}

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
	case !hasSlaveRow && config.IS_PRIMARY:
		status.Role = "master"
	case !ioRunning || !sqlRunning:
		status.Role = "broken"
	default:
		status.Role = "replica"
	}

	return status, nil
}
