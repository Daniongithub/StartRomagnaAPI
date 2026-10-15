package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
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

	primary, err := isPrimary(DB_STATIC)
	config.IS_PRIMARY.Store(primary)
	if err != nil {
		fmt.Println("InitStatic errore determinazione DB principale:", err)
	}
}

func InitRT() {
	DB_RT, err = newDB(config.DB_HOST, config.DB_PORT, config.DB_USERNAME, config.DB_PASSWORD, "start_gtfs_rt")
	if err != nil {
		fmt.Println("InitRT errore di connessione al database del GTFS dinamico:", err)
	}

	primary, err := isPrimary(DB_RT)
	config.IS_PRIMARY.Store(primary)
	if err != nil {
		fmt.Println("InitRT errore determinazione DB principale:", err)
	}
}

func InitMezzi() {
	DB_MEZZI, err = newDB(config.DB_HOST, config.DB_PORT, config.DB_USERNAME, config.DB_PASSWORD, "ertpl_mezzi")
	if err != nil {
		fmt.Println("InitMezzi errore di connessione al database del GTFS dinamico:", err)
	}

	primary, err := isPrimary(DB_MEZZI)
	config.IS_PRIMARY.Store(primary)
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
	case !hasSlaveRow && !status.ReadOnly:
		status.Role = "master"
	case !hasSlaveRow && status.ReadOnly:
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
	config.SCHEDULER.Shutdown()

	db := DB_STATIC
	if _, err := db.Exec("SET GLOBAL read_only = 1"); err != nil {
		return fmt.Errorf("set read_only: %w", err)
	}

	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return fmt.Errorf("show slave status: %w", err)
	}
	wasSlave := rows.Next()
	rows.Close()

	if wasSlave {
		if _, err := db.Exec("STOP SLAVE"); err != nil {
			return fmt.Errorf("stop slave: %w", err)
		}
		if _, err := db.Exec("RESET SLAVE ALL"); err != nil {
			return fmt.Errorf("reset slave: %w", err)
		}
	}
	if _, err := db.Exec("SET GLOBAL gtid_slave_pos = ''"); err != nil {
		return fmt.Errorf("reset gtid_slave_pos: %w", err)
	}

	config.IS_PRIMARY.Store(false)
	return nil
}

func RunResync(masterHost string) {
	if err := HaltLocal(); err != nil {
		log.Printf("resync: halt locale fallito: %v", err)
		return
	}

	log.Printf("resync: avvio dump da %s", masterHost)

	dump := exec.Command("mariadb-dump",
		"-h", masterHost,
		"-u", config.DB_USERNAME,
		"--single-transaction", // snapshot consistente InnoDB, nessun lock prolungato
		"--master-data=1",      // include la posizione di partenza come SQL nel dump
		"--gtid",               // ...in forma GTID (SET GLOBAL gtid_slave_pos), non file/offset
		"--routines", "--triggers",
		"--databases", "start_gtfs_static", "start_gtfs_rt", // <-- conferma i nomi veri
	)
	// password via env, non come argomento -p: altrimenti finisce in chiaro
	// nella process list (`ps aux`) visibile a chiunque abbia accesso alla macchina
	dump.Env = append(os.Environ(), "MYSQL_PWD="+config.DB_PASSWORD)

	restore := exec.Command("mariadb", "-h", "127.0.0.1", "-u", config.DB_USERNAME)
	restore.Env = append(os.Environ(), "MYSQL_PWD="+config.DB_PASSWORD)

	pipe, err := dump.StdoutPipe()
	if err != nil {
		log.Printf("resync: errore creazione pipe: %v", err)
		return
	}
	restore.Stdin = pipe

	if err := restore.Start(); err != nil {
		log.Printf("resync: impossibile avviare l'import: %v", err)
		return
	}
	if err := dump.Run(); err != nil {
		log.Printf("resync: dump fallito: %v", err)
		return
	}
	if err := restore.Wait(); err != nil {
		log.Printf("resync: import fallito: %v", err)
		return
	}

	log.Println("resync: dump+import completati, configuro la replica")

	db := DB_STATIC
	if _, err := db.Exec(
		"CHANGE MASTER TO MASTER_HOST=?, MASTER_PORT=3306, MASTER_USER=?, MASTER_PASSWORD=?, MASTER_USE_GTID=slave_pos",
		masterHost, config.REPLICA_USERNAME, config.REPLICA_PASSWORD,
	); err != nil {
		log.Printf("resync: change master fallito: %v", err)
		return
	}
	if _, err := db.Exec("START SLAVE"); err != nil {
		log.Printf("resync: start slave fallito: %v", err)
		return
	}

	log.Println("resync: completato, replica attiva da", masterHost)
}
