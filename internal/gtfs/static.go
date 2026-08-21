package gtfs

import (
	"StartRomagnaAPI/config"
	"StartRomagnaAPI/internal/auth"
	"StartRomagnaAPI/internal/repository"
	"fmt"
	"net/http"

	"github.com/Leocraft1/gtfsparser-with-reader"
)

func UpdateStatic() {
	urlRA := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Ravenna.zip"
	urlFC := config.START_GTFS_ROOT + "/AVM/GTFSStatic_FOCE.zip"
	urlRN := config.START_GTFS_ROOT + "/AVM/GTFSStatic_Rimini.zip"

	feedRA, err := getStaticFeed(urlRA, true)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}
	feedFC, err := getStaticFeed(urlFC, false)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}
	feedRN, err := getStaticFeed(urlRN, false)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}
	
	repository.SaveTrips(feedRA, feedFC, feedRN)
	repository.SaveCalDates(feedRA, feedFC, feedRN)
}


func getStaticFeed(url string, skip_valid bool) (*gtfsparserwr.Feed, error) {
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