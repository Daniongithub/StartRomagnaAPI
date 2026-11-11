package gtfs

import (
	"fmt"
	"log"
	"net/http"

	// "time" // Uncomment with the duration metrics below.

	"startromagnaapi/config"
	"startromagnaapi/internal/auth"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/static"
	"startromagnaapi/internal/service"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
	"github.com/jmoiron/sqlx"
)

func UpdateStatic() error {
	return repository.WithAdvisoryLock(repository.DB_STATIC, "startromagnaapi:update-static-gtfs", updateStatic)
}

func updateStatic() error {
	// start := time.Now() // Uncomment to measure import duration.
	log.Println("Started static GTFS update.")
	feeds := make(map[string]*gtfsparserwr.Feed, 3)
	for _, source := range []struct {
		basin, file    string
		skipValidation bool
	}{
		{"RA", "GTFSStatic_Ravenna.zip", true}, {"FC", "GTFSStatic_FOCE.zip", false}, {"RN", "GTFSStatic_Rimini.zip", false},
	} {
		feed, err := getStatic(config.START_GTFS_ROOT+"/AVM/"+source.file, source.skipValidation)
		if err != nil {
			return fmt.Errorf("Static %s: %w", source.basin, err)
		}
		feeds[source.basin] = feed
	}
	if err := static.SaveSnapshot(feeds); err != nil {
		return err
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

func getStatic(url string, skip_valid bool) (*gtfsparserwr.Feed, error) {
	req, err := auth.BasicAuth("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := auth.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Static feed: HTTP %s", resp.Status)
	}

	feed := gtfsparserwr.NewFeed()
	// Skip stop-time validation for feeds with known data issues.
	if skip_valid {
		feed.Opts.SkipStopTimeValidation = true
	}

	if err := feed.ParseReader(resp.Body); err != nil {
		return nil, fmt.Errorf("GetStaticFeed: parse gtfs: %w", err)
	}

	return feed, nil
}
