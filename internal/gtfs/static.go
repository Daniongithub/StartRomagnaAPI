package gtfs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	// "time" // Uncomment with the duration metrics below.

	"startromagnaapi/config"
	"startromagnaapi/internal/auth"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/service"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
	"github.com/jmoiron/sqlx"
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
		feed, err := getStatic(config.START_GTFS_ROOT+"/AVM/"+source.file, source.skipValidation)
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

func getStatic(feedURL string, skipValidation bool) (*gtfsparserwr.Feed, error) {
	req, err := auth.BasicAuth(http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Invalid static feed URL")
	}

	resp, err := auth.Client.Do(req)
	if err != nil {
		var requestErr *url.Error
		if errors.As(err, &requestErr) {
			err = requestErr.Err
		}
		return nil, fmt.Errorf("Download static feed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Static feed: HTTP %d", resp.StatusCode)
	}

	feed := gtfsparserwr.NewFeed()
	// Skip stop-time validation for feeds with known data issues.
	if skipValidation {
		feed.Opts.SkipStopTimeValidation = true
	}

	if err := parseStaticFeed(feed, resp.Body); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidStaticFeed, err)
	}

	return feed, nil
}

func parseStaticFeed(feed *gtfsparserwr.Feed, body io.Reader) (err error) {
	// The parser can panic while reading a malformed CSV header.
	defer func() {
		if failure := recover(); failure != nil {
			err = fmt.Errorf("Static GTFS parser panic: %v", failure)
		}
	}()
	return feed.ParseReader(body)
}
