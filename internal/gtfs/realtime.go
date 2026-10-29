package gtfs

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"startromagnaapi/config"
	"startromagnaapi/internal/auth"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/realtime"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

const (
	serviceAlerts    string = "service-alerts"
	tripUpdates      string = "trip-updates"
	vehiclePositions string = "vehicle-positions"
)

var basins []string = []string{"RA", "FC", "RN"}

// Fetch every basin before touching the DB: a failed download preserves the snapshot.
func fetchRT(kind string) (map[string]*gtfs.FeedMessage, error) {
	type result struct {
		feed *gtfs.FeedMessage
		err  error
	}
	results := make([]result, len(basins))
	var wg sync.WaitGroup
	for i, basin := range basins {
		wg.Go(func() { results[i].feed, results[i].err = getRT(rtURL(basin, kind)) })
	}
	wg.Wait()
	feeds := make(map[string]*gtfs.FeedMessage, len(basins))
	for i, basin := range basins {
		if results[i].err != nil {
			return nil, fmt.Errorf("Fetch %s feed for %s: %w", kind, basin, results[i].err)
		}
		feeds[basin] = results[i].feed
	}
	return feeds, nil
}

func UpdateAlerts() error {
	feeds, err := fetchRT(serviceAlerts)
	if err != nil {
		return err
	}
	return realtime.SaveAlerts(feeds)
}

func UpdateTripUpdates() error {
	feeds, err := fetchRT(tripUpdates)
	if err != nil {
		return err
	}
	return repository.InTransaction(repository.DB_RT, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec("DELETE FROM stop_time_updates"); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM trip_updates"); err != nil {
			return err
		}
		if err := realtime.SaveTripUpdates(tx, feeds); err != nil {
			return err
		}
		return realtime.SaveStopTimeUpd(tx, feeds)
	})
}

func UpdateVehiclePositions() error {
	feeds, err := fetchRT(vehiclePositions)
	if err != nil {
		return err
	}
	return repository.InTransaction(repository.DB_RT, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec("DELETE FROM vehicle_positions"); err != nil {
			return err
		}
		return realtime.SaveVehiclePositions(tx, feeds)
	})
}

func rtURL(basin, ft string) string {
	return config.START_GTFS_RT_ROOT + "/start-gtfs-rt-" + ft + "-" + strings.ToLower(basin) + ".pb"
}

func getRT(url string) (*gtfs.FeedMessage, error) {
	req, err := auth.BasicAuth(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Create request: %w", err)
	}

	resp, err := auth.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Request %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Request %s: server returned %s",
			url,
			resp.Status,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read %s: %w", url, err)
	}

	var feed gtfs.FeedMessage

	if err := proto.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("Unmarshal %s: %w", url, err)
	}

	if feed.Header == nil {
		return nil, fmt.Errorf("Feed %s: missing header", url)
	}
	if feed.GetHeader().GetIncrementality() != gtfs.FeedHeader_FULL_DATASET {
		return nil, fmt.Errorf("Feed %s: differential feeds are unsupported", url)
	}
	return &feed, nil
}
