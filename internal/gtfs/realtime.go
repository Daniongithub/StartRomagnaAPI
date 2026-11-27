package gtfs

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"startromagnaapi/config"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/realtime"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/jlaffaye/ftp"
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
	return "/start/gtfs-rt/start-gtfs-rt-" + ft + "-" + strings.ToLower(basin) + ".pb"
}

func getRT(path string) (*gtfs.FeedMessage, error) {
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
	zipped, err := conn.Retr(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to open realtime GTFS file")
	}

	body, err := io.ReadAll(zipped)
	if err != nil {
		return nil, fmt.Errorf("Read %s: %w", path, err)
	}

	var feed gtfs.FeedMessage

	if err := proto.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("Unmarshal %s: %w", path, err)
	}

	if feed.Header == nil {
		return nil, fmt.Errorf("Feed %s: missing header", path)
	}
	if feed.GetHeader().GetIncrementality() != gtfs.FeedHeader_FULL_DATASET {
		return nil, fmt.Errorf("Feed %s: differential feeds are unsupported", path)
	}
	return &feed, nil
}
