package gtfs

import (
	"startromagnaapi/config"
	"startromagnaapi/internal/auth"
	"startromagnaapi/internal/repository/static"
	"fmt"
	"net/http"
	"time"

	gtfsparserwr "github.com/Leocraft1/gtfsparser-with-reader"
)

func UpdateStatic() {
	fmt.Println("Updating static GTFS...")
	start := time.Now()
	urlRA := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Ravenna.zip"
	urlFC := config.START_GTFS_ROOT + "/AVM/GTFSStatic_FOCE.zip"
	urlRN := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Rimini.zip"

	feedRA, err := getStatic(urlRA, true)
	if err != nil {
		fmt.Println("UpdateStatic error reading feed:", err)
	}
	feedFC, err := getStatic(urlFC, false)
	if err != nil {
		fmt.Println("UpdateStatic error reading feed:", err)
	}
	feedRN, err := getStatic(urlRN, false)
	if err != nil {
		fmt.Println("UpdateStatic error reading feed:", err)
	}

	static.SaveCalDates(feedRA, feedFC, feedRN)
	static.SaveRoutes(feedRA, feedFC, feedRN)
	static.SaveShapes(feedRA, feedFC, feedRN)
	static.SaveStopTimes(feedRA, feedFC, feedRN)
	static.SaveStops(feedRA, feedFC, feedRN)
	static.SaveTrips(feedRA, feedFC, feedRN)

	elapsed := time.Since(start)
	fmt.Printf("Updated static GTFS. Elapsed: %d min %d sec\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
}

func getStatic(url string, skip_valid bool) (*gtfsparserwr.Feed, error) {
	req, err := auth.BasicAuth("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	feed := gtfsparserwr.NewFeed()
	//If there are known problems with stop times, this bypasses control
	if skip_valid {
		feed.Opts.SkipStopTimeValidation = true
	}

	if err := feed.ParseReader(resp.Body); err != nil {
		return nil, fmt.Errorf("GetStaticFeed: parse gtfs: %w", err)
	}

	return feed, nil
}
