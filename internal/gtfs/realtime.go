package gtfs

/*func UpdateRT() {
	fmt.Println("Updating realtime GTFS...")
	start := time.Now()
	urlRA_SA := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-service-alerts-ra.pb"
	urlRA_TU := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-trip-updates-ra.pb"
	urlRA_VP := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-vehicle-positions-ra.pb"

	urlFC_SA := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-service-alerts-fc.pb"
	urlFC_TU := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-trip-updates-fc.pb"
	urlFC_VP := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-vehicle-positions-fc.pb"

	urlRN_SA := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-service-alerts-rn.pb"
	urlRN_TU := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-trip-updates-rn.pb"
	urlRN_VP := config.START_GTFS_RT_ROOT + "/start-gtfs-rt-vehicle-positions-rn.pb"

	feedRA, err := getStatic(urlRA, true)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}
	feedFC, err := getStatic(urlFC, false)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}
	feedRN, err := getStatic(urlRN, false)
	if err != nil {
		fmt.Println("TripsHandler error reading feed:", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Updated realtime GTFS. Elapsed: %d min %d sec\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
}

func getRT(url string, skip_valid bool) (*gtfsparserwr.Feed, error) {
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
}*/
