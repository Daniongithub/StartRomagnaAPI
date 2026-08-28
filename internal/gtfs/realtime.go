package gtfs

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/auth"
	"StartRomagnaAPI/internal/model"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

const (
	ServiceAlerts    model.RTFeedType = "service-alerts"
	TripUpdates      model.RTFeedType = "trip-updates"
	VehiclePositions model.RTFeedType = "vehicle-positions"
)

func UpdateRT() {
	fmt.Println("Updating realtime GTFS...")
	start := time.Now()

	ctx := context.Background()

	type result struct {
		area string
		typ  model.RTFeedType
		feed *gtfs.FeedMessage
	}

	jobs := make(chan struct {
		area string
		typ  model.RTFeedType
	})

	results := make(chan result, 9)

	var wg sync.WaitGroup

	// Massimo 3 richieste contemporaneamente.
	for range 3 {

		wg.Go(func() {

			for job := range jobs {
				url := rtURL(job.area, job.typ)

				feed, err := getRT(ctx, url)
				if err != nil {
					fmt.Printf(
						"Realtime GTFS error [%s/%s]: %v\n",
						job.area,
						job.typ,
						err,
					)
					continue
				}

				results <- result{
					area: job.area,
					typ:  job.typ,
					feed: feed,
				}
			}
		})
	}

	for _, area := range []string{"ra", "fc", "rn"} {
		for _, typ := range []model.RTFeedType{
			ServiceAlerts,
			TripUpdates,
			VehiclePositions,
		} {
			jobs <- struct {
				area string
				typ  model.RTFeedType
			}{
				area: area,
				typ:  typ,
			}
		}
	}

	close(jobs)

	wg.Wait()
	close(results)

	// Qui fai il parsing/normalizzazione.
	for result := range results {
		fmt.Printf(
			"Loaded %s/%s: %d entities\n",
			result.area,
			result.typ,
			len(result.feed.Entity),
		)
	}

	elapsed := time.Since(start)
	fmt.Printf(
		"Updated realtime GTFS. Elapsed: %d min %d sec\n",
		int(elapsed.Minutes()),
		int(elapsed.Seconds())%60,
	)
}

func rtURL(area string, feedType model.RTFeedType) string {
	return fmt.Sprintf(
		"%s/start-gtfs-rt-%s-%s.pb",
		config.START_GTFS_RT_ROOT,
		string(feedType),
		area,
	)
}

func getRT(ctx context.Context, url string) (*gtfs.FeedMessage, error) {
	req, err := auth.BasicAuth(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req = req.WithContext(ctx)

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
