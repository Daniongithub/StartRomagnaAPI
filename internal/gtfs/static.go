package gtfs

import (
	"errors"
	"fmt"
	"log"
	"strings"

	// "time" // Uncomment with the duration metrics below.

	"startromagnaapi/config"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/service"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
	"github.com/jmoiron/sqlx"
	"github.com/jlaffaye/ftp"
)

var errInvalidStaticFeed = errors.New("Invalid static GTFS feed")

// StaticImportError retains detailed errors for logs and a summary safe for notifications.
type StaticImportError struct {
	cause   error
	summary string
}

func (e *StaticImportError) Error() string   { return e.cause.Error() }
func (e *StaticImportError) Unwrap() error   { return e.cause }
func (e *StaticImportError) Summary() string { return e.summary }

func UpdateStatic() error {
	return repository.WithAdvisoryLock(repository.DB_STATIC, "startromagnaapi:update-static-gtfs", updateStatic)
}

func updateStatic() error {
	// start := time.Now() // Uncomment to measure import duration.
	log.Println("Started static GTFS update.")
	var failures []error
	var summaries []string
	for _, source := range []struct {
		basin, file    string
		skipValidation bool
	}{
		{"RA", "GTFSStatic_Ravenna.zip", true}, {"FC", "GTFSStatic_FOCE.zip", false}, {"RN", "GTFSStatic_Rimini.zip", false},
	} {
		feed, err := getStatic("/start/gtfs/AVM/"+source.file, source.skipValidation)
		stage := "Download failed"
		if err == nil {
			stage = "Database import failed"
			err = static.SaveSnapshot(source.basin, feed)
		} else if errors.Is(err, errInvalidStaticFeed) {
			stage = "Feed parsing or validation failed"
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("Static %s: %w", source.basin, err))
			summaries = append(summaries, source.basin+": "+stage)
			continue
		}
		log.Printf("Updated static GTFS for %s", source.basin)
	}
	if err := errors.Join(failures...); err != nil {
		return &StaticImportError{cause: err, summary: strings.Join(summaries, "\n")}
	}
	// log.Printf("Static GTFS updated in %s", time.Since(start))
	return nil
}

func UpdateRouteInfos() error {
	raw, err := service.ScrapeRouteInfo(config.ROUTE_INFO_URL)
	if err != nil {
		return err
	}
	infos := service.ProcessRouteInfo(*raw)
	if len(infos) == 0 {
		return fmt.Errorf("Route info: no recognized routes, preserving existing data")
	}
	return repository.InTransaction(repository.DB_STATIC, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec("DELETE FROM route_infos"); err != nil {
			return err
		}
		return static.SaveRouteInfos(tx, infos)
	})
}

func getStatic(feedPath string, skipValidation bool) (*gtfsparserwr.Feed, error) {
	//FTP connection
    conn, err := ftp.Dial(config.START_GTFS_FTP)
	if err != nil {
		log.Fatalf("Failed to create FTP client: %s", err)
	}
	//Login to FTP
	err = conn.Login(config.FTP_USERNAME, config.FTP_PASSWORD)
	if err != nil {
		log.Fatalf("Failed to login to FTP: %s", err)
	}
    defer conn.Quit()
	zipped, err := conn.Retr(feedPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open static GTFS file")
	}

	feed := gtfsparserwr.NewFeed()

	if err := feed.ParseReader(zipped); err != nil {
		return nil, fmt.Errorf("GetStaticFeed: parse gtfs: %w", err)
	}

	return feed, nil 
}
