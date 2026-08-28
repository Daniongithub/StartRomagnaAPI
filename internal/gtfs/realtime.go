package gtfs

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/auth"
	"StartRomagnaAPI/internal/repository/realtime"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

const (
	serviceAlerts    string   = "service-alerts"
	tripUpdates      string   = "trip-updates"
	vehiclePositions string   = "vehicle-positions"
)

var basins []string = []string{"RA", "FC", "RN"}

func UpdateAlerts() {
	var feeds = make(map[string]*gtfs.FeedMessage)
	for _, val := range basins {
		url := rtURL(val, serviceAlerts)

		rt, err := getRT(url)
		if err != nil {
			fmt.Println("UpdateAlerts error parsing feed")
		}

		feeds[val] = rt
	}

	realtime.SaveAlerts(feeds)
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
