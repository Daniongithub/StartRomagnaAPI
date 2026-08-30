package gtfs

import (
	"fmt"
	"io"
	"net/http"
	"startromagnaapi/config"
	"startromagnaapi/internal/auth"
	"startromagnaapi/internal/repository"
	"startromagnaapi/internal/repository/realtime"
	"strings"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

const (
	serviceAlerts    string = "service-alerts"
	tripUpdates      string = "trip-updates"
	vehiclePositions string = "vehicle-positions"
)

var basins []string = []string{"RA", "FC", "RN"}

func UpdateAlerts() {
	start := time.Now()
	feeds := make(map[string]*gtfs.FeedMessage)
	alertsMap := make(map[string]map[string]bool)
	tmp := make(map[string]bool)

	alertsRA := realtime.GetServiceAlertsBasin("RA")
	for _, val := range alertsRA {
		tmp[val.Id] = true
	}
	alertsMap["RA"] = tmp
	alertsFC := realtime.GetServiceAlertsBasin("FC")
	for _, val := range alertsFC {
		tmp[val.Id] = true

	}
	alertsMap["FC"] = tmp
	alertsRN := realtime.GetServiceAlertsBasin("RN")
	for _, val := range alertsRN {
		tmp[val.Id] = true
	}
	alertsMap["RN"] = tmp

	for _, val := range basins {
		url := rtURL(val, serviceAlerts)

		rt, err := getRT(url)
		if err != nil {
			fmt.Println("UpdateAlerts error parsing feed")
		}

		feeds[val] = rt
	}

	realtime.SaveAlerts(feeds, alertsMap)
	elapsed := time.Since(start)
	fmt.Printf("Updated ServiceAlerts. Elapsed: %d min %d sec\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
}

func UpdateTripUpdates() {
	start := time.Now()
	feeds := make(map[string]*gtfs.FeedMessage)

	for _, val := range basins {
		url := rtURL(val, tripUpdates)

		rt, err := getRT(url)
		if err != nil {
			fmt.Println("UpdateTripUpdates error parsing feed")
		}

		feeds[val] = rt
	}

	repository.StartTransactionTU()
	realtime.DeleteAllStopTimeUpd()
	realtime.DeleteAllTripUpdates()
	realtime.SaveTripUpdates(feeds)
	realtime.SaveStopTimeUpd(feeds)
	repository.CommitTransactionTU()

	elapsed := time.Since(start)
	fmt.Printf("Updated TripUpdates. Elapsed: %d min %d sec\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
}

func UpdateVehiclePositions() {
	start := time.Now()
	feeds := make(map[string]*gtfs.FeedMessage)

	for _, val := range basins {
		url := rtURL(val, vehiclePositions)

		rt, err := getRT(url)
		if err != nil {
			fmt.Println("UpdateVehiclePositions error parsing feed")
		}

		feeds[val] = rt
	}

	repository.StartTransactionVP()
	realtime.DeleteAllVehiclePositions()
	realtime.SaveVehiclePositions(feeds)
	repository.CommitTransactionVP()

	elapsed := time.Since(start)
	fmt.Printf("Updated UpdateVehiclePositions. Elapsed: %d min %d sec\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
}

func rtURL(basin, ft string) string {
	return config.START_GTFS_RT_ROOT + "/start-gtfs-rt-" + ft + "-" + strings.ToLower(basin) + ".pb"
}

func getRT(url string) (*gtfs.FeedMessage, error) {
	req, err := auth.BasicAuth(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"request %s: server returned %s",
			url,
			resp.Status,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}

	var feed gtfs.FeedMessage

	if err := proto.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", url, err)
	}

	return &feed, nil
}
