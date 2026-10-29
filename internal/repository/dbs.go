package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
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
	DB_ORCH   *sqlx.DB
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

func IsPrimary(db *sqlx.DB) error {
	tx, err := db.Begin()
	var varName, uptimeStr string
	if err := db.QueryRow("SHOW GLOBAL STATUS LIKE 'Uptime'").Scan(&varName, &uptimeStr); err != nil {
		return fmt.Errorf("Failed to read database uptime: %w", err)
	}
	uptime, _ := strconv.Atoi(uptimeStr)
	if uptime < 60 {
		// MariaDB has just started or restarted: Fence residual privileges
		// and let the arbiter decide the role on its next tick.
		return HaltLocal()
	}

	if err != nil {
		return fmt.Errorf("Failed to begin write-probe transaction: %w", err)
	}
	defer tx.Rollback() // The probe does not need to persist any changes.

	_, err = tx.Exec("UPDATE arbiter_probe SET touched_at = NOW() WHERE id = 1")

	if err == nil {
		config.IS_PRIMARY.Store(true)
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1142 { // Command denied.
		config.IS_PRIMARY.Store(false)
		return nil
	}

	return fmt.Errorf("Write probe failed unexpectedly: %w", err)
}

func initDB(name, user, password string) *sqlx.DB {
	db, err := newDB(config.DB_HOST, config.DB_PORT, user, password, name)
	if err != nil {
		log.Fatalf("Failed to connect to database %s: %v", name, err)
	}
	return db
}

func InitStatic() {
	DB_STATIC = initDB("start_gtfs_static", config.DB_USERNAME, config.DB_PASSWORD)
	if err := IsPrimary(DB_STATIC); err != nil {
		log.Printf("Failed to determine the primary database: %v", err)
	}
}

func InitRT()    { DB_RT = initDB("start_gtfs_rt", config.DB_USERNAME, config.DB_PASSWORD) }
func InitMezzi() { DB_MEZZI = initDB("ertpl_mezzi", config.DB_USERNAME, config.DB_PASSWORD) }

func InitOrchestrator() {
	DB_ORCH = initDB("start_gtfs_static", config.DB_ORCH_USERNAME, config.DB_ORCH_PASSWORD)
}

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

	// Scan dynamically because SHOW SLAVE STATUS has many columns
	// and its output differs between MariaDB versions.
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
	case !hasSlaveRow && config.IS_PRIMARY.Load():
		status.Role = "master"
	case !hasSlaveRow && !config.IS_PRIMARY.Load():
		status.Role = "unconfigured"
	case !ioRunning || !sqlRunning:
		status.Role = "broken"
	case lag == nil || *lag > 0:
		status.Role = "resyncing"
	default:
		status.Role = "replica"
	}

	return status, nil
}

func HaltLocal() error {

	db := DB_STATIC
	if _, err := db.Exec("SET GLOBAL read_only = 1"); err != nil {
		return fmt.Errorf("Set read_only: %w", err)
	}

	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return fmt.Errorf("Show slave status: %w", err)
	}
	wasSlave := rows.Next()
	rows.Close()

	if wasSlave {
		if _, err := db.Exec("STOP SLAVE"); err != nil {
			return fmt.Errorf("Stop slave: %w", err)
		}
		if _, err := db.Exec("RESET SLAVE ALL"); err != nil {
			return fmt.Errorf("Reset slave: %w", err)
		}
	}
	if _, err := db.Exec("SET GLOBAL gtid_slave_pos = ''"); err != nil {
		return fmt.Errorf("Reset gtid_slave_pos: %w", err)
	}

	config.IS_PRIMARY.Store(false)
	return nil
}

func RunResync(masterHost string) {
	if err := HaltLocal(); err != nil {
		log.Printf("Resync: Failed to halt the local node: %v", err)
		return
	}

	log.Printf("Resync: Starting database dump from %s", masterHost)

	dump := exec.Command("mariadb-dump",
		"-h", masterHost,
		"-u", config.DB_USERNAME,
		"--single-transaction", // Consistent InnoDB snapshot without long-lived locks.
		"--master-data=1",      // Include the starting replication position as SQL in the dump.
		"--gtid",               // Use GTID (SET GLOBAL gtid_slave_pos) instead of a file offset.
		"--routines", "--triggers",
		"--databases", "start_gtfs_static", "start_gtfs_rt", // Databases to include in the dump.
	)
	// Pass the password through the environment instead of a -p argument
	// to avoid exposing it in the process list (`ps aux`).
	dump.Env = append(os.Environ(), "MYSQL_PWD="+config.DB_PASSWORD)

	restore := exec.Command("mariadb", "-h", "127.0.0.1", "-u", config.DB_USERNAME)
	restore.Env = append(os.Environ(), "MYSQL_PWD="+config.DB_PASSWORD)

	pipe, err := dump.StdoutPipe()
	if err != nil {
		log.Printf("Resync: Failed to create the dump pipe: %v", err)
		return
	}
	restore.Stdin = pipe

	if err := restore.Start(); err != nil {
		log.Printf("Resync: Failed to start the import: %v", err)
		return
	}
	if err := dump.Run(); err != nil {
		log.Printf("Resync: Database dump failed: %v", err)
		return
	}
	if err := restore.Wait(); err != nil {
		log.Printf("Resync: Database import failed: %v", err)
		return
	}

	log.Println("Resync: Dump and import completed; configuring replication")

	db := DB_STATIC
	if _, err := db.Exec(
		"CHANGE MASTER TO MASTER_HOST=?, MASTER_PORT=3306, MASTER_USER=?, MASTER_PASSWORD=?, MASTER_USE_GTID=slave_pos",
		masterHost, config.REPLICA_USERNAME, config.REPLICA_PASSWORD,
	); err != nil {
		log.Printf("Resync: Failed to change the master: %v", err)
		return
	}
	if _, err := db.Exec("START SLAVE"); err != nil {
		log.Printf("Resync: Failed to start replication: %v", err)
		return
	}

	log.Println("Resync: Completed; replicating from", masterHost)
}
